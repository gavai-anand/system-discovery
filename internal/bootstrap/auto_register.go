package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"system-discovery/internal/app/constants"
	"time"

	"system-discovery/internal/dto"
)

func AutoRegister(app *App) {
	self := os.Getenv("SELF")

	var request dto.RegisterRequest
	request.NodeID = self

	app.RegistrationService.Register(context.Background(), request)

	peersEnv := os.Getenv("PEERS")
	if peersEnv == "" {
		return
	}

	peers := strings.Split(peersEnv, ",")
	peers = append(peers, self)
	for _, peer := range peers {
		if peer == "" || peer == self {
			continue
		}

		waitForPeer(app, peer)
		// Step 1: Register
		endpoint := fmt.Sprintf(constants.REGISTER_API, peer)
		var registerRequest dto.RegisterRequest
		registerRequest.NodeID = self

		fmt.Println("registering peer ", peer, endpoint)
		resp, _, err := app.ServiceCall.Post(context.Background(), endpoint, registerRequest, nil)
		if err != nil {
			log.Println("Register failed:", err)
			continue
		}
		var response dto.PeersListResponse
		endpoint = fmt.Sprintf(constants.PEER_LIST_API, peer)
		resp, _, err = app.ServiceCall.Get(context.Background(), endpoint, nil)
		if err != nil {
			log.Println("Getting Peer failed:", err)
			continue
		}

		// Step 2: Get peer list
		if err := json.Unmarshal(resp, &response); err != nil {
			log.Println("Failed to parse peer list:", err)
			continue
		}
		fmt.Println("peer list:", response.Peers)
		peerList := response.Peers
		// Step 3: Update local store
		app.DiscoveryService.SyncAndStorePeers(context.Background(), dto.SyncPeersRequest{
			Peers: peerList,
		})

		// Step 4: Broadcast self
		for _, p := range peerList {
			if p == self {
				continue
			}
			endpoint = fmt.Sprintf(constants.SYNC_PEERS_API, p)
			_, _, err := app.ServiceCall.Post(context.Background(), endpoint, dto.SyncPeersRequest{Peers: []string{self}}, nil)
			if err != nil {
				fmt.Println("Failed to sync peer:", err)
				continue
			}

			endpoint = fmt.Sprintf(constants.OPERATIONS_API, p)

			resp, _, err := app.ServiceCall.Get(context.Background(), endpoint, nil)
			if err != nil {
				fmt.Println("Failed to get value:", err)
				continue
			}
			fmt.Println("operations:", string(resp))
			var operationResponse dto.OperationsResponse
			if err := json.Unmarshal(resp, &operationResponse); err != nil {
				fmt.Println("Failed to parse value:", err)
				continue
			}
			data := operationResponse.Data
			for _, op := range data {
				app.CounterService.Replicate(context.Background(), op)
			}
		}
	}
}

func waitForPeer(app *App, peer string) {
	for {
		_, cancel := context.WithTimeout(context.Background(), 1*time.Second)

		// try a simple GET (health endpoint preferred)
		endpoint := fmt.Sprintf(constants.HEARTBEAT_API, peer)
		_, _, err := app.ServiceCall.Get(context.Background(), endpoint, nil, 2*time.Second)

		cancel()

		if err == nil {
			log.Println("Peer is ready:", peer)
			return
		}

		log.Println("Waiting for peer:", peer)
		time.Sleep(1 * time.Second)
	}
}

package elastosadenine

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cyber-republic/go-grpc-adenine/v2/elastosadenine/stubs/node_rpc"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type NodeRpc struct {
	Connection *grpc.ClientConn
}

type JWTInfoNodeRpc struct {
	Network string      `json:"network"`
	Chain   string      `json:"chain"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func NewNodeRpc(host string, port int, production bool) *NodeRpc {
	address := fmt.Sprintf("%s:%d", host, port)
	var opts []grpc.DialOption
	if production == false {
		opts = []grpc.DialOption{
			grpc.WithInsecure(),
		}
	} else {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(config)),
		}
	}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	return &NodeRpc{Connection: conn}
}

func (n *NodeRpc) Close() {
	n.Connection.Close()
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentCrcCouncil(apiKey, did, network string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	params["state"] = "all"
	return n.RpcMethod(apiKey, did, network, "mainchain", "listcurrentcrs", params).(map[string]interface{})
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentCrcCandidates(apiKey, did, network string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	params["start"] = 0
	params["state"] = "all"
	return n.RpcMethod(apiKey, did, network, "mainchain", "listcrcandidates", params).(map[string]interface{})
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentDposSupernodes(apiKey, did, network string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	params["start"] = 0
	params["state"] = "all"
	return n.RpcMethod(apiKey, did, network, "mainchain", "listproducers", params).(map[string]interface{})
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentArbitratorGroup(apiKey, did, network string) map[string]interface{} {
	height := n.GetCurrentHeight(apiKey, did, network, "mainchain")
	return n.GetArbitratorGroup(apiKey, did, network, height)
}

// Common method for mainchain only
func (n *NodeRpc) GetArbitratorGroup(apiKey, did, network, height string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	params["height"] = height
	return n.RpcMethod(apiKey, did, network, "mainchain", "getarbitratorgroupbyheight", params).(map[string]interface{})
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentArbitratorsInfo(apiKey, did, network string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	return n.RpcMethod(apiKey, did, network, "mainchain", "getarbitersinfo", params).(map[string]interface{})
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentBlockConfirm(apiKey, did, network string) map[string]interface{} {
	height := n.GetCurrentHeight(apiKey, did, network, "mainchain")
	return n.GetBlockConfirm(apiKey, did, network, height)
}

// Common method for mainchain only
func (n *NodeRpc) GetBlockConfirm(apiKey, did, network, height string) map[string]interface{} {
	h, err := strconv.Atoi(height)
	if err != nil {
		log.Fatalf("Failed to execute 'GetBlockConfirm' method: %v", err)
	}
	params := make(map[string]interface{}, 0)
	params["height"] = h
	params["verbosity"] = 1
	return n.RpcMethod(apiKey, did, network, "mainchain", "getconfirmbyheight", params).(map[string]interface{})
}

// Common method for mainchain only
func (n *NodeRpc) GetCurrentMiningInfo(apiKey, did, network string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	return n.RpcMethod(apiKey, did, network, "mainchain", "getmininginfo", params).(map[string]interface{})
}

// Common method for mainchain, did sidechain, token and eth sidechain
func (n *NodeRpc) GetCurrentBlockInfo(apiKey, did, network, chain string) map[string]interface{} {
	height := n.GetCurrentHeight(apiKey, did, network, chain)
	return n.GetBlockInfo(apiKey, did, network, chain, height)
}

// Common method for mainchain, did sidechain, token and eth sidechain
func (n *NodeRpc) GetBlockInfo(apiKey, did, network, chain, height string) map[string]interface{} {
	if chain == "eth" {
		params := []interface{}{}
		heightInt, _ := strconv.Atoi(height)
		params = append(params, fmt.Sprintf("0x%02x", heightInt))
		params = append(params, true)
		return n.RpcMethod(apiKey, did, network, chain, "eth_getBlockByNumber", params).(map[string]interface{})
	} else {
		params := make(map[string]interface{}, 0)
		params["height"] = height
		return n.RpcMethod(apiKey, did, network, chain, "getblockbyheight", params).(map[string]interface{})
	}
}

// Common method for mainchain, did sidechain, token and eth sidechain
func (n *NodeRpc) GetCurrentBalance(apiKey, did, network, chain, address string) interface{} {
	var currentBalance interface{}

	if chain == "eth" {
		params := []string{address, "latest"}
		currentBalanceHex := n.RpcMethod(apiKey, did, network, chain, "eth_getBalance", params).(string)
		currentBalanceHexCleaned := strings.Replace(currentBalanceHex, "0x", "", -1)
		currentBalanceInt64, _ := strconv.ParseInt(currentBalanceHexCleaned, 16, 64)
		currentBalance = fmt.Sprintf("%v", currentBalanceInt64)
	} else {
		params := make(map[string]interface{}, 0)
		params["address"] = address
		if chain == "token" {
			result := n.RpcMethod(apiKey, did, network, chain, "getreceivedbyaddress", params).(map[string]interface{})
			balance := make(map[string]string)
			for key, value := range result {
				balance[key] = value.(string)
			}
			currentBalance = balance
		} else {
			currentBalance = n.RpcMethod(apiKey, did, network, chain, "getreceivedbyaddress", params).(string)
		}
	}
	return currentBalance
}

// Common method for mainchain, did sidechain, token and eth sidechain
func (n *NodeRpc) GetCurrentHeight(apiKey, did, network, chain string) string {
	var currentHeight string
	params := make(map[string]interface{}, 0)
	if chain == "eth" {
		currentHeightHex := n.RpcMethod(apiKey, did, network, chain, "eth_blockNumber", params).(string)
		currentHeightHexCleaned := strings.Replace(currentHeightHex, "0x", "", -1)
		currentHeightInt64, _ := strconv.ParseInt(currentHeightHexCleaned, 16, 64)
		currentHeight = fmt.Sprintf("%v", currentHeightInt64)
	} else {
		nodeState := n.GetCurrentNodeState(apiKey, did, network, chain)
		currentHeight = fmt.Sprintf("%.0f", nodeState["height"].(float64))
	}
	return currentHeight
}

// Common method for mainchain, did sidechain and token sidechain
func (n *NodeRpc) GetCurrentNodeState(apiKey, did, network, chain string) map[string]interface{} {
	params := make(map[string]interface{}, 0)
	return n.RpcMethod(apiKey, did, network, chain, "getnodestate", params).(map[string]interface{})
}

func (n *NodeRpc) RpcMethod(apiKey, did, network, chain, method string, params interface{}) interface{} {
	var data interface{}
	client := node_rpc.NewNodeRpcClient(n.Connection)

	jwtInfo, _ := json.Marshal(JWTInfoNodeRpc{
		Network: network,
		Chain:   chain,
		Method:  method,
		Params:  params,
	})

	claims := JWTClaim{
		JwtInfo: string(jwtInfo),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(tokenExpiration).Unix(),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err := jwtToken.SignedString([]byte(apiKey))
	if err != nil {
		log.Fatalf("Failed to execute 'RpcMethod' method: %v", err)
	}
	md := metadata.Pairs("did", did)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := client.RpcMethod(ctx, &node_rpc.Request{Input: jwtTokenString})
	if err != nil {
		log.Fatalf("Failed to execute 'RpcMethod' method: %v", err)
	}

	if response.Status == true {
		recvToken, err := jwt.Parse(response.Output, func(recvToken *jwt.Token) (interface{}, error) {
			if _, ok := recvToken.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", recvToken.Header["alg"])
			}
			return []byte(apiKey), nil
		})

		if recvClaims, ok := recvToken.Claims.(jwt.MapClaims); ok && recvToken.Valid {
			strMap := recvClaims["jwt_info"].(map[string]interface{})
			data = strMap["result"].(interface{})
		} else {
			log.Fatalf("Failed to execute 'RpcMethod' method: %v", err)
		}
	}
	return data
}

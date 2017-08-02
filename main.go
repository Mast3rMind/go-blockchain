package main

import(
    "fmt"
    "flag"
    "net"
    "strings"
    "regexp"
)

/*------------------------*
 * COMMAND LINE INTERFACE *
 *------------------------*/

func main() {
    // set up flags
    var listenPort string
    flag.StringVar(&listenPort, "l", "1999", "")
    flag.StringVar(&listenPort, "listen", "1999", "")

    var seedPort string
    flag.StringVar(&seedPort, "s", "", "")
    flag.StringVar(&seedPort, "seed", "", "")

    var helpFlag bool
    flag.BoolVar(&helpFlag, "h", false, "")
    flag.BoolVar(&helpFlag, "help", false, "")

    var joinFlag bool
    flag.BoolVar(&joinFlag, "j", false, "")
    flag.BoolVar(&joinFlag, "join", false, "")

    setFlag(flag.CommandLine)
    flag.Parse()

    listenPort = ":" + listenPort 
    seedPort = ":" + seedPort

    if helpFlag {
        showGlobalHelp()
        return
    }

    fmt.Println(".................................")
    if listenPort != ":" {
        fmt.Printf("Listen port:                %s \n", listenPort)
    }
    if seedPort != ":" {
        fmt.Printf("Seed port:                  %s \n", seedPort)
    }
    if (joinFlag && seedPort != ""){
        fmt.Printf("Will attempt to join network\n")
    }
    fmt.Println(".................................\n")

    // create channels
    inputChannel            := make(chan string)
    transmissionChannel     := make(chan *Transmission)
    connChannel             := make(chan net.Conn)
    disconnChannel          := make(chan net.Conn)
    requestChannel          := make(chan net.Conn)
    addressesChannel        := make(chan []string)
    blockChannel            := make(chan Block)
    blockchainRequestChannel:= make(chan net.Conn)
    blockchainChannel       := make(chan Blockchain)

    // create node    
    myNode := Node{make(map[net.Conn]int), 0, Blockchain{[]Block{genesisBlock}}, "", "", map[string]bool{}}
    myNode.updateAddress(listenPort)
    myNode.updateSeed(seedPort)

    startListening(listenPort, connChannel, inputChannel)
    if joinFlag { // if the user requested to join a seed node // need to make sure you can't join if you don't supply a seed
        fmt.Println("Dialing seed node at port " + seedPort + "...")
         go dialNode(seedPort, connChannel)
    }

    // handle go routines
    for {
        select {
            case conn    := <- connChannel: // listener picked up new conn
                myNode.incrementConnID()
                myNode.connections[conn] = myNode.nextConnID // assign connection an ID
                go listenToConn(conn, transmissionChannel, disconnChannel, requestChannel, addressesChannel, blockchainRequestChannel, blockchainChannel)

            case disconn := <- disconnChannel: // established connection disconnected
                connID := myNode.connections[disconn]
                delete(myNode.connections, disconn) // remove the connection from the nodes list of connections
                fmt.Printf("* Connection %v has been disconnected \n", connID)

            case trans := <- transmissionChannel:  // new transmission sent to node
                notMinedAndValid   := myNode.seenBlocks[string(trans.Block.Hash)] == false  && trans.BeenSent == true && myNode.blockchain.isValidBlock(trans.Block)
                notMinedAndInvalid := myNode.seenBlocks[string(trans.Block.Hash)] == false  && trans.BeenSent == true && !myNode.blockchain.isValidBlock(trans.Block)
                minedButNotSent    := myNode.seenBlocks[string(trans.Block.Hash)] == true   && trans.BeenSent == false
                alreadySent        := myNode.seenBlocks[string(trans.Block.Hash)] == true   && trans.BeenSent == true
                if notMinedAndValid {
                    myNode.seenBlocks[string(trans.Block.Hash)] = true
                    myNode.blockchain.addBlock(trans.Block)
                    fmt.Printf("[notMinedAndValid] Added block #%v sent from network to my blockchain, and sending it to network\n", trans.Block.Index)
                    trans.updateSender(myNode.address)
                    forwardTransToNetwork(*trans, myNode.connections) // forward messages to the rest of network
                } else if notMinedAndInvalid { // not mined and not valid
                    myNode.seenBlocks[string(trans.Block.Hash)] = true
                    myBlockchainLength := myNode.blockchain.getLastBlock().Index
                    if trans.Block.Index > myBlockchainLength {
                        connThatSentHigherBlockIndex := myNode.getConnForAddress(trans.Sender)
                        fmt.Println("I was sent a block with a higher index, now requesting full chain to validate")
                        requestBlockchain(connThatSentHigherBlockIndex)
                    }
                    fmt.Printf("[notMinedAndInvalid] Did not add block #%v sent from network to my chain, did not forward\n", trans.Block.Index)
                } else if minedButNotSent { //mined but not sent out yet,
                    trans.updateBeenSent()
                    trans.updateSender(myNode.address) 
                    fmt.Printf("[minedButNotSent] Mined block #%v, sending to network\n", trans.Block.Index)
                    forwardTransToNetwork(*trans, myNode.connections) // forward messages to the rest of network
                } else if alreadySent{
                    fmt.Printf("[alreadySent] Already seen block #%v, did not forward", trans.Block.Index)
                } else {
                    fmt.Println("Some other case, this should not occur:")
                    // fmt.Println("myNode.seenBlocks[string(trans.Block.Hash)]")
                    // fmt.Println(myNode.seenBlocks[string(trans.Block.Hash)])
                    // fmt.Println("trans.BeenSent")
                    // fmt.Println((trans.BeenSent))
                    // fmt.Println("myNode.blockchain.isValidBlock(trans.Block)")
                    // fmt.Println(myNode.blockchain.isValidBlock(trans.Block))
                }

            case conn := <-  requestChannel:  // was requested addresses to send
                addressesToSendTo := myNode.getRemoteAddresses()
                sendConnectionsToNode(conn, addressesToSendTo)

            case addresses := <- addressesChannel:  //received addresses to add
                fmt.Print("Seed node sent these addresses to connect to:\n-> " )
                fmt.Println(addresses)
                approvedAddresses := []string{}
                for i := range addresses {
                    r, _ := regexp.Compile(":.*") // match everything after the colon
                    port := r.FindString(addresses[i])
                    if len(port) == 5 {  // in a real network this should simply be 1999
                        go dialNode(port, connChannel)
                        approvedAddresses = append(approvedAddresses, addresses[i])
                    }
                }
                fmt.Print("These new connections will be added:\n->")
                fmt.Println(approvedAddresses)

            case conn    := <- blockchainRequestChannel:
                sendBlockchainToNode(conn, myNode.blockchain)

            case blockchain := <- blockchainChannel:
                fmt.Println("You were sent a blockchain")
                if blockchain.isValidChain() {
                    myNode.blockchain = blockchain
                    fmt.Println("Blockchain accepted: ")
                    fmt.Println(blockchain)
                } else {
                    fmt.Println("Blockchain rejected, invalid")
                }

            case block   := <- blockChannel:
                if myNode.blockchain.isValidBlock(block){
                    myNode.blockchain.addBlock(block)
                    myNode.seenBlocks[string(block.Hash)] = true // specify weve now seen this block but don't update the trans address until its processed there
                    go sendTransFromMinedBlock(block, transmissionChannel)
                } else {
                    fmt.Printf("Did not add mined block #%v\n", block.Index)
                }
                go myNode.blockchain.mineBlock(blockChannel)

            case input   := <- inputChannel: // user entered some input
                outgoingArgs := strings.Fields(strings.Split(input,"\n")[0]) // remove newline char and seperate into array by whitespace
                arg0 := strings.ToLower(outgoingArgs[0])
                switch arg0 {
                case "mine":
                    go myNode.blockchain.mineBlock(blockChannel)                        
                case "getchain":
                    if myNode.seed == "" {
                        fmt.Println("You must have a seed node to request a blockchain")
                    } else{
                        seedConn := myNode.getConnForAddress(myNode.seed)
                        requestBlockchain(seedConn)                        
                    }
                case "getconns":
                    if myNode.hasConnectionOfAddress(myNode.seed){
                        seedConn := myNode.getConnForAddress(myNode.seed)
                        fmt.Println("Requesting more connections from seed " + myNode.seed + " ...")
                        requestConnections(seedConn)
                    } else {
                        fmt.Println("You are not connected to your seed node to make a request..")
                    }
                case "node":
                    myNode.printNode()
                case "help":
                    showNodeHelp()
                default:
                    fmt.Println("Enter 'help' for options.")
                }
        }

    }
}

/*-------------------*
 *   HELP & SETUP    *
 *-------------------*/

func showGlobalHelp() {
fmt.Println(
`NAME:
   go-blockchain - blockchain network

USAGE:
   go-blockchain [global options]

COMMANDS:
   go-blockchain      launches a node

GLOBAL OPTIONS:
    -l, --listen     assigns the listening port for the server        (default = 1999).
    -s, --seed       assigns the port of the seed                     (default = 2000).
    -j, --join       attempt to join the network                      (default = false).
    -h, --help       prints this help info

NODE COMMANDS:
    send      sends the subsequent text to the network
    getconns   requests the list of nodes from your seed node and attempts to connect to each
    getchain  requests seed node for their version of the blockchain
    node      prints the information associated with your node
    help      prints the node command help info`)
}

func showNodeHelp(){
fmt.Println(
`
NODE COMMANDS:
    send      sends the subsequent text to the network
    getconns  requests the list of nodes from your seed node and attempts to connect to each
    node      prints the information associated with your node
    help      prints the node command help info`)
}

func setFlag(flag *flag.FlagSet) {
    flag.Usage = func() {
        showGlobalHelp()
    }
}

# sendingnetwork-bot

A Golang SDN client.

## Install

```sh
go get github.com/sending-network/sendingnetwork-bot
```

## Usage

### Prepare a configuration file
Provide server endpoint, wallet address and private key in config.yaml:
```yaml
endpoint: ""
wallet_address: ""
private_key: ""
```
You can use an existing wallet account, or generate a new account by running:
```shell
go run tools/generate_wallet_account.go
```

### Create an instance of `Client`
After reading the configuration file, create an instance of `Client`
```go
package main

import (
	"os"
	
	sdnclient "github.com/sending-network/sendingnetwork-bot"
	"gopkg.in/yaml.v3"
)

func main() {
    configData, _ := os.ReadFile("config.yaml")
    config := sdnclient.Config{}
    _ = yaml.Unmarshal(configData, &config)
    cli, _ := sdnclient.NewClient(&config)
}
```

### Call API functions
```go
// create new room
respCreateRoom, _ := cli.CreateRoom(&sdnclient.ReqCreateRoom{
	Name:   "TestRoom",
})

// invite user to the room
respInviteUser, _ := cli.InviteUser(respCreateRoom.RoomID, &sdnclient.ReqInviteUser{
    UserID: userID,
})

// logout to invalidate access token
respLogout, _ := cli.Logout()
```

## Examples
See more use cases in `examples` directory.

## License
Apache 2.0
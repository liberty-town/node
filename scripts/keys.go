package main

import (
	"encoding/base64"
	"fmt"
	"liberty-town/node/addresses"
)

func main() {

	privateKey := addresses.GenerateNewPrivateKey()

	fmt.Println("PRIVATE KEY", base64.StdEncoding.EncodeToString(privateKey.Key))
	fmt.Println("PUBLIC KEY", base64.StdEncoding.EncodeToString(privateKey.GeneratePublicKey()))

	fmt.Println("PRIVATE KEY WIF", base64.StdEncoding.EncodeToString(privateKey.Serialize()))

}

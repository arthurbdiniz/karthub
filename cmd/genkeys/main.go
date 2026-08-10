package main

import (
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		panic(err)
	}
	fmt.Printf("KARTHUB_VAPID_PUBLIC_KEY=%s\n", pub)
	fmt.Printf("KARTHUB_VAPID_PRIVATE_KEY=%s\n", priv)
}

package main

import (
		"fmt"
		b64 "encoding/base64"
		"bufio"
		"os"
		"encoding/json"
		"strings"
		"unicode/utf8"
		"flag"
		)


func main() {

	decodeFlag := flag.Bool("decode", false, "Decode Mode")
	flag.Parse()

	if *decodeFlag {
		fmt.Println("Decode token with JSESSIONLOCK.")
		
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter JSESSIONLOCK: ")
		j_lock, _ := reader.ReadString('\n')

		fmt.Print("Enter token: ")
		token, _ := reader.ReadString('\n')

		j_lock = strings.TrimSuffix(j_lock, "\r\n")
		token = strings.TrimSuffix(token, "\r\n")

		dec := decode(token, j_lock)

		result := make(map[string]interface{})
		json.Unmarshal([]byte(dec), &result)

		fmt.Print("Login: ")
		fmt.Println(result["l"])
		fmt.Print("Password: ")
		fmt.Println(result["p"])
		return
			
	} 

	fmt.Println("Encode data with JSESSIONLOCK")
		
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter JSESSIONLOCK: ")
	j_lock, _ := reader.ReadString('\n')

	fmt.Print("Enter login: ")
	login, _ := reader.ReadString('\n')

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')

	fmt.Print("Enter code (press enter to default value): ")
	code, _ := reader.ReadString('\n')

	j_lock = strings.TrimSuffix(j_lock, "\r\n")
	login = strings.TrimSuffix(login, "\r\n")
	password = strings.TrimSuffix(password, "\r\n")
	code = strings.TrimSuffix(code, "\r\n")
		
	if len(code) == 0 {
		code = "A"
	}

	token := map[string]string{"l":login, "p":password}
	json_token, _ := json.Marshal(token)
	z := encode(string(json_token), j_lock, code)	
	fmt.Print("Fake token: ")
	fmt.Println(z)
} 

func encode(h string, e string, b string) string {
	// var c = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// fuck random, fixed charcode
	//b := "U"
	g, _ := utf8.DecodeRune([]byte(b))
	//g := int()
	a := ""
	e = e[g:]
	for d := 0; d< len(h); d++ {
		a = a + string(h[d] ^ e[d % len(e)])
		
	}
	return string(b) + string(b64.StdEncoding.EncodeToString([]byte(a)))
}

func decode(t string, e string) string {
	b := t[0]
	xored_base := t[1:len(t)]
	h, _ := b64.StdEncoding.DecodeString(xored_base)
	//g, _ := utf8.DecodeRune(b)
	e = e[b:]
	a := ""
	for i := 0; i < len(h); i++ {
		a = a + string(h[i] ^ e[i % len(e)])
		
	}
	return a
}
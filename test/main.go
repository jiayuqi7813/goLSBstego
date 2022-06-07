package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"stego.go"
)

func main(){
	encode()
	decode()

}
func decode(){
	input,_:=os.Open("out_file.png")
	defer input.Close()
	reader := bufio.NewReader(input)
	img,_:=png.Decode(reader)
	sizeOfMessge := stego.GetMessageSizeFromImage(img)
	message :=stego.Decode(sizeOfMessge,img)
	fmt.Println(string(message))
}
func encode(){
	input,_:=os.Open("or.png")
	message, _ := ioutil.ReadFile("tes.txt")
	reader := bufio.NewReader(input)
	img,_:=png.Decode(reader)
	w := new(bytes.Buffer)
	err := stego.Encode(w,img,message)
	if err != nil {
		log.Printf("Error Encoding file %v", err)
		return
	}
	outFile, _ := os.Create("out_file.png") // create file
	w.WriteTo(outFile) // write buffer to it
	outFile.Close()
}

package stego

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

func EncodeNRGBA(writeBuffer *bytes.Buffer, rgbImg *image.NRGBA,messgae []byte)error {
	var messageLen = uint32(len(messgae))
	var width = rgbImg.Bounds().Dx() 	//返回宽度
	var height = rgbImg.Bounds().Dy()	//返回高度
	var co color.NRGBA
	var bit byte
	var isbool bool
	if MaxEncodeSize(rgbImg) < messageLen +4 {
		return errors.New("message is larger than image")
	}
	a,b,c,d := splitToBytes(messageLen)
	messgae = append([]byte{d},messgae...)
	messgae = append([]byte{c},messgae...)
	messgae = append([]byte{b},messgae...)
	messgae = append([]byte{a},messgae...)
	ch := make(chan byte,100)
	go getNextBitFromString(messgae,ch)

	for x := 0; x < width; x++ {
		for y :=0;y<height;y++ {
			co = rgbImg.NRGBAAt(x, y)//获取某个像素的颜色

			//R
			bit,isbool = <-ch
			if !isbool{
				//bit不足
				rgbImg.SetNRGBA(x,y,co)
				png.Encode(writeBuffer,rgbImg)
			}
			setLsb(&co.R,bit)
			//G
			bit,isbool = <-ch
			if !isbool{
				rgbImg.SetNRGBA(x,y,co)
				png.Encode(writeBuffer,rgbImg)
				return nil
			}
			setLsb(&co.G,bit)
			//B
			bit,isbool = <-ch
			if !isbool{
				rgbImg.SetNRGBA(x,y,co)
				png.Encode(writeBuffer,rgbImg)
				return nil
			}
			setLsb(&co.B,bit)


			rgbImg.SetNRGBA(x,y,co)

		}
	}
	err := png.Encode(writeBuffer,rgbImg)
	fmt.Println("err")
	return err
}



//Encode 将图片中的信息编码到图片中
//最小图片大小是23像素
//封装，image.Image自动转换成image.NRGBA
/*
	Input:
		writeBuffer *bytes.Buffer : 写入的图片
		message []byte : 写入的信息
		picInputFile image.Image : 图像数据加密
 */
func Encode(writeBuffer *bytes.Buffer,picInputfile image.Image, message []byte) error {
	rgbImg := imageToNRGBA(picInputfile)
	return EncodeNRGBA(writeBuffer,rgbImg,message)
}


func decodeNRGBA(startOffset uint32,msgLen uint32,rgbImg *image.NRGBA)(message []byte){
	var byteIndex uint32
	var bitIndex uint32
	width := rgbImg.Bounds().Dx()
	height := rgbImg.Bounds().Dy()
	var co color.NRGBA
	var lsb byte
	message = append(message, 0)
	for x := 0; x < width; x++ {
		for y:=0;y<height;y++ {

			co = rgbImg.NRGBAAt(x, y)
			//R
			lsb = getLSB(co.R)
			message[byteIndex] = setBitInByte(message[byteIndex],bitIndex,lsb)
			bitIndex++

			if bitIndex > 7 {
				bitIndex = 0
				byteIndex++
				if byteIndex >= msgLen+startOffset {
					return message[startOffset : msgLen+startOffset]
				}

				message = append(message, 0)
			}
			//G
			lsb = getLSB(co.G)
			message[byteIndex] = setBitInByte(message[byteIndex],bitIndex,lsb)
			bitIndex++
			if bitIndex	> 7 {
				bitIndex = 0
				byteIndex++

				if byteIndex >= msgLen+startOffset {
					return message[startOffset : msgLen+startOffset]
				}
				message = append(message, 0)

			}
			//B
			lsb = getLSB(co.B)
			message[byteIndex] = setBitInByte(message[byteIndex],bitIndex,lsb)
			bitIndex++
			if bitIndex > 7 {
				bitIndex = 0
				byteIndex++

				if byteIndex >= msgLen+startOffset {
					return message[startOffset : msgLen+startOffset]
				}

				message = append(message, 0)
			}

		}
	}
	return
}

func decode(startOffset uint32, msgLen uint32, pictureInputFile image.Image) (message []byte) {

	rgbImage := imageToNRGBA(pictureInputFile)
	return decodeNRGBA(startOffset, msgLen, rgbImage)
}

func Decode(msgLen uint32, pictureInputFile image.Image) (message []byte) {
	return decode(4, msgLen, pictureInputFile) // the offset of 4 skips the "header" where message length is defined

}

func GetMessageSizeFromImage(pictureInputFile image.Image) (size uint32) {

	sizeAsByteArray := decode(0, 4, pictureInputFile)
	size = combineToInt(sizeAsByteArray[0], sizeAsByteArray[1], sizeAsByteArray[2], sizeAsByteArray[3])
	return
}

func combineToInt(one, two, three, four byte) (ret uint32) {
	ret = uint32(one)
	ret = ret << 8
	ret = ret | uint32(two)
	ret = ret << 8
	ret = ret | uint32(three)
	ret = ret << 8
	ret = ret | uint32(four)
	return
}

func getLSB(b byte) byte {
	if b%2 == 0 {
		return 0
	}
	return 1
}

func setBitInByte(b byte, indexInByte uint32, bit byte) byte {
	var mask byte = 0x80
	mask = mask >> uint(indexInByte)

	if bit == 0 {
		mask = ^mask
		b = b & mask
	} else if bit == 1 {
		b = b | mask
	}
	return b
}


// MaxEncodeSize 找出最低有效编码存储字节数量
//((width * height * 3) / 8 ) - 4
// 结果至少是4
func MaxEncodeSize(img image.Image) uint32 {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	eval := ((width * height * 3) / 8) - 4
	if eval < 4 {
		eval = 0
	}
	return uint32(eval)
}

//splitToBytes 将一个数字分割成一个字节的数组
func splitToBytes(num uint32) (one,two,three,four byte) {
	one = byte(num>>24)
	var mask uint32 = 255
	two = byte((num>>16) & mask)
	three = byte((num>>8) & mask)
	four = byte(num & mask)
	return
}

//getBitFromByte 获取一个字节的某一位
func getBitFromByte(b byte,indexInByte int)byte{
	b = b << uint(indexInByte)
	var mask byte = 0x80
	var bit = mask & b
	if bit ==128{
		return 1
	}
	return 0
}

//getNextBitFromString 每次获取字符串下一个字节
func getNextBitFromString(ArrayByte []byte,ch chan byte) {
	var offsetInByte  int
	var offsetInBitsIntoByte int
	var choiceByte byte
	stringLen := len(ArrayByte)
	for {
		if offsetInByte >=stringLen{
			close(ch)
			return
		}
		choiceByte = ArrayByte[offsetInByte]
		ch <- getBitFromByte(choiceByte, offsetInBitsIntoByte)
		offsetInBitsIntoByte++
		if offsetInBitsIntoByte >=8{
			offsetInBitsIntoByte = 0
			offsetInByte++
		}

	}
}


//setLsb 设置最低有效位 (true is 1, false is 0)
func setLsb(b *byte, bit byte) {
	if bit ==1{
		*b = *b|1
	}else if bit == 0{
		var mask byte = 0xFE
		*b = *b & mask
	}

}


// imageToNRGBA converts image.Image to image.NRGBA
func imageToNRGBA(src image.Image) *image.NRGBA {
	bounds := src.Bounds()

	var m *image.NRGBA
	var width, height int

	width = bounds.Dx()
	height = bounds.Dy()

	m = image.NewNRGBA(image.Rect(0, 0, width, height))

	draw.Draw(m, m.Bounds(), src, bounds.Min, draw.Src)
	return m
}
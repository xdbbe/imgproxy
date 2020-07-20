package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/h2non/bimg"
	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

func conv(inputFilename string, outputExtDesired string, c *fiber.Ctx) ([]byte, bool) {
	inputImg := doRequest("https://cdn.xdb.be/img/" + inputFilename)

	outputWidth, _ := strconv.Atoi(c.FormValue("width"))
	outputHeight, _ := strconv.Atoi(c.FormValue("height"))

	size, err := bimg.Size(inputImg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	outputType := imageType(outputExtDesired, inputImg)

	if outputWidth == 0 {
		outputWidth = size.Width
	}

	if outputHeight == 0 {
		outputHeight = size.Height
	}

	opts := bimg.Options{
		Type:   outputType,
		Width:  outputWidth,
		Height: outputHeight,
	}

	// resize and transcode image
	if !checkCachePost(c) {
		outputImg, err := bimg.NewImage(inputImg).Process(opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return outputImg, true
	}
	return inputImg, false
}

func doRequest(url string) []byte {
	if x, found := ca.Get(url); found {
		return x.([]byte)
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)

	fasthttp.Do(req, resp)
	ca.Set(url, resp.Body(), cache.DefaultExpiration)
	return resp.Body()
}

func checkCachePost(c *fiber.Ctx) bool {
	extension := c.Params("extension")
	hash := c.Params("hash")

	if x, found := ca.Get(strings.Join([]string{hash, extension, c.FormValue("width"), c.FormValue("height")}, ";")); found {
		resp := x.([]byte)
		c.Set("content-type", "image/"+extension)
		c.Send(resp)
		c.Set("X-Powered-By", "xdb-imgproxy")
		fmt.Println("postcache")
		return true
	}
	return false
}

func imageType(name string, inputImg []byte) bimg.ImageType {
	switch strings.ToLower(name) {
	case "jpeg":
		return bimg.JPEG
	case "png":
		return bimg.PNG
	case "webp":
		return bimg.WEBP
	case "tiff":
		return bimg.TIFF
	case "gif":
		return bimg.GIF
	case "svg":
		return bimg.SVG
	case "pdf":
		return bimg.PDF
	default:
		return bimg.DetermineImageType(inputImg)
	}
}

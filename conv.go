package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/discordapp/lilliput"
	"github.com/gofiber/fiber"
	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

var EncodeOptions = map[string]map[int]int{
	".jpeg": map[int]int{lilliput.JpegQuality: 100},
	".png":  map[int]int{lilliput.PngCompression: 0},
	".webp": map[int]int{lilliput.WebpQuality: 100},
}

var ops = lilliput.NewImageOps(8192)
var m sync.Mutex

func conv(inputFilename string, outputExtension string, c *fiber.Ctx) ([]byte, bool) {
	defer ops.Clear()
	inputBuf := doRequest("https://cdn.xdb.be/img/" + inputFilename)
	decoder, err := lilliput.NewDecoder(inputBuf)
	if err != nil {
		return inputBuf, false
	}
	defer decoder.Close()
	outputWidth, _ := strconv.Atoi(c.FormValue("width"))
	outputHeight, _ := strconv.Atoi(c.FormValue("height"))

	header, err := decoder.Header()
	// create a buffer to store the output image, 50MB in this case
	outputImg := make([]byte, 50*1024*1024)

	// use user supplied filename to guess output type if provided
	// otherwise don't transcode (use existing type)
	outputType := "." + strings.ToLower(decoder.Description())
	if outputExtension != "" {
		outputType = outputExtension
	}

	if outputWidth == 0 {
		outputWidth = header.Width()
	}

	if outputHeight == 0 {
		outputHeight = header.Height()
	}

	resizeMethod := lilliput.ImageOpsFit

	if outputWidth == header.Width() && outputHeight == header.Height() {
		resizeMethod = lilliput.ImageOpsNoResize
	}

	opts := &lilliput.ImageOptions{
		FileType:             outputType,
		Width:                outputWidth,
		Height:               outputHeight,
		ResizeMethod:         resizeMethod,
		NormalizeOrientation: true,
		EncodeOptions:        EncodeOptions[outputType],
	}

	// resize and transcode image
	m.Lock()
	if !checkCachePost(c) {
		outputImg, err = ops.Transform(decoder, opts, outputImg)
		if err != nil {
			fmt.Printf("error transforming image, %s\n", err)
		}
	}
	m.Unlock()
	return outputImg, true
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
		return true
	}
	return false
}

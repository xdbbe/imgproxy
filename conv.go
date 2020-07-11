package main

import (
	"fmt"
	"strings"

	"github.com/discordapp/lilliput"
	"github.com/valyala/fasthttp"
)

var EncodeOptions = map[string]map[int]int{
	".jpeg": map[int]int{lilliput.JpegQuality: 100},
	".png":  map[int]int{lilliput.PngCompression: 0},
	".webp": map[int]int{lilliput.WebpQuality: 100},
}

var ops = lilliput.NewImageOps(8192)

func conv(inputFilename string, outputWidth int, outputHeight int, outputExtension string) ([]byte, bool) {
	inputBuf := doRequest("https://s3.nl-ams.scw.cloud/cdn.xdb.be/img/" + inputFilename)
	decoder, err := lilliput.NewDecoder(inputBuf)
	if err != nil {
		return inputBuf, false
	}
	defer decoder.Close()

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
	outputImg, err = ops.Transform(decoder, opts, outputImg)
	if err != nil {
		fmt.Printf("error transforming image, %s\n", err)
	}

	return outputImg, true
}

func doRequest(url string) []byte {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)

	fasthttp.Do(req, resp)

	return resp.Body()
}

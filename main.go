package main

import (
	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
)

func main() {
	//Memory Profiler, disable in prod
	//defer profile.Start(profile.MemProfile).Stop()

	app := fiber.New(&fiber.Settings{
		//enable if there is enough ram (min 32-64 gig)
		//Prefork:          true,
		DisableKeepalive: true,
	})

	//SSL, enable if standalone with own ip
	// cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// config := &tls.Config{Certificates: []tls.Certificate{cer}}

	//Recover, idk if this actually does anything
	app.Use(middleware.Recover())
	//Logger
	f, _ := os.Create("imgproxy.log")
	defer f.Close()
	w := bufio.NewWriter(f)
	go func() {
		for {
			w.Flush()
			time.Sleep(1 * time.Second)
		}
	}()
	app.Use(middleware.Logger(w))
	//Main Logic
	app.Get("/img/:hash.:extension", func(c *fiber.Ctx) {
		hash := c.Params("hash")
		extension := c.Params("extension")
		width, _ := strconv.Atoi(c.FormValue("width"))
		height, _ := strconv.Atoi(c.FormValue("height"))
		resp, success := conv(hash, width, height, hash+"."+extension)
		c.Send(resp)
		if success == true {
			c.Set("content-type", "image/"+extension)
		} else {
			c.Set("content-type", "application/xml")
		}
		c.Set("X-Powered-By", "xdb-imgproxy")
	})

	app.Get("*", func(c *fiber.Ctx) {
		c.Status(404)
		c.Set("X-Powered-By", "xdb-imgproxy")
	})
	app.Listen(8080 /* , config */)
}

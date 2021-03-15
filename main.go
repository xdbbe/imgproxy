package main

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patrickmn/go-cache"
)

var ca = cache.New(5*time.Minute, 10*time.Minute)

func main() {
	//Memory Profiler, disable in prod
	//defer profile.Start(profile.MemProfile).Stop()

	app := fiber.New(fiber.Config{
		DisableKeepalive: true,
	})

	app.Use(recover.New())
	//Logger
	file, err := os.OpenFile("./imgproxy.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
	}
	defer file.Close()
	app.Use(logger.New(logger.Config{
		Output: file,
	}))
	//Main Logic
	app.Get("/img/:hash.:extension", func(c *fiber.Ctx) error {
		extension := c.Params("extension")
		hash := c.Params("hash")

		if x, found := ca.Get(strings.Join([]string{hash, extension, c.FormValue("width"), c.FormValue("height")}, ";")); found {
			resp := x.([]byte)
			c.Set("content-type", "image/"+extension)
			return c.Send(resp)
			c.Set("X-Powered-By", "xdb-imgproxy")
		}else{
			resp, success := conv(hash, extension, c)
			if success == true {
				ca.Set(strings.Join([]string{hash, extension, c.FormValue("width"), c.FormValue("height")}, ";"), resp, cache.DefaultExpiration)
			}
			if resp != nil {
				return c.Send(resp)
			}
		}
		return fiber.NewError(500)
	})

	app.Get("*", func(c *fiber.Ctx) error {
		return fiber.NewError(404)
	})
	app.Listen(":8080")
}

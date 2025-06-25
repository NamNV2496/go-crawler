package controller

import (
	"fmt"
	"log/slog"
	"reflect"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/namnv2496/seo/configs"
	"github.com/namnv2496/seo/pkg/validate"
)

func Start(
	urlController IController,
) (*echo.Echo, error) {
	conf := configs.LoadConfig()
	e := newEchoServer()
	publicGroup := e.Group("/api/v1/public")
	// CRUD method
	publicGroup.POST("/url", wrapReponse(urlController.CreateNewUrl))
	publicGroup.PUT("/url/:id", wrapReponse(urlController.UpdateUrl))
	publicGroup.GET("/url", wrapReponse(urlController.GetUrl))
	publicGroup.GET("/urls", wrapReponse(urlController.GetUrls))
	// build and parse url
	publicGroup.POST("/url/build", wrapReponse(urlController.BuildUrl))
	publicGroup.POST("/url/parse", wrapReponse(urlController.ParseUrl))
	publicGroup.POST("/url/dynamic_keyword", wrapReponse(urlController.DynamicParamParseByUrl))
	// sitemap and robots.txt
	publicGroup.GET("/sitemap", wrapReponse(urlController.Sitemap))
	publicGroup.GET("/robots.txt", wrapReponse(urlController.Robots))

	if err := e.Start(fmt.Sprintf(":%s", conf.AppPort)); err != nil {
		e.Logger.Fatal(err)
	}
	slog.Info("Server is running on port: ", "port", conf.AppPort)
	return e, nil
}

func newEchoServer() *echo.Echo {
	e := echo.New()
	e.Validator = validate.NewValidator()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	return e
}

func wrapReponse(function any) echo.HandlerFunc {
	ftype := reflect.TypeOf(function)
	fval := reflect.ValueOf(function)
	// if ftype.NumIn() != 2 {
	// 	panic("function must have 2 parameters")
	// }
	if fval.Kind() != reflect.Func {
		panic("function must be a function")
	}
	runtime.FuncForPC(fval.Pointer()).Name()
	errorIndex := ftype.NumOut() - 1

	return func(c echo.Context) error {
		// execute function
		// req := reflect.New(ftype.In(1))
		// if err := c.Bind(req.Interface()); err != nil {
		// 	return err
		// }
		// err := c.Validate(req.Interface())
		// if err != nil {
		// 	return err
		// }
		// res := fval.Call([]reflect.Value{
		// 	reflect.ValueOf(c),
		// 	req.Elem(),
		// })
		var args []reflect.Value
		args = append(args, reflect.ValueOf(c))
		if ftype.NumIn() > 1 {
			req := reflect.New(ftype.In(1))
			if err := c.Bind(req.Interface()); err != nil {
				return err
			}
			err := c.Validate(req.Interface())
			if err != nil {
				return err
			}
			args = append(args, req.Elem())
		}

		// Call the handler function
		res := fval.Call(args)
		if !res[errorIndex].IsNil() {
			slog.Error("handler error", "error", res[errorIndex].Interface().(error))
			return res[errorIndex].Interface().(error)
		}
		resp := c.Response()
		output := res[0].Interface()
		slog.Info("output: ", output)
		return c.JSON(resp.Status, output)
	}
}

type {{ $.InterfaceName }} interface {
{{range .MethodSet}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{end}}
}

type {{$.Name}} struct{
		Server {{ $.InterfaceName }}
		Router gin.IRouter
		Resp  interface {
			Error(ctx *gin.Context, err error)
			ParamsError (ctx *gin.Context, err error)
			Success(ctx *gin.Context, data interface{})
		}
}

// Resp 返回值
type Default{{$.Name}}Resp struct {}

// Error 返回错误信息
func (resp Default{{$.Name}}Resp) Error(ctx *gin.Context, err error) {
	code := -1
	status := 500
	msg := "未知错误"
	
	if err == nil {
		msg += ", err is nil"
		ctx.JSON(status, map[string]interface{}{
			"code": code,
			"msg":  msg,
		})
		return
	}

	type iCode interface{
		HTTPCode() int
		Message() string
		Code() int
	}

	var c iCode
	if errors.As(err, &c) {
		status = c.HTTPCode()
		code = c.Code()
		msg = c.Message()
	}

	_ = ctx.Error(err)

	ctx.JSON(status, map[string]interface{}{
		"code": code,
		"msg":  msg,
	})
}

// ParamsError 参数错误
func (resp Default{{$.Name}}Resp) ParamsError (ctx *gin.Context, err error) {
	_ = ctx.Error(err)
	ctx.JSON(400, map[string]interface{}{
		"code": 400,
		"msg":  "参数错误",
	})
}

// Success 返回成功信息
func (resp Default{{$.Name}}Resp) Success(ctx *gin.Context, data interface{}) {
	// resp需要定义code，为了让code和数据平级
	ctx.JSON(200, data)
}


{{range .Methods}}
func (s *{{$.Name}}) {{ .HandlerName }} (ctx *gin.Context) {
	var in {{.Request}}
{{if .HasPathParams }}
	if err := ctx.ShouldBindUri(&in); err != nil {
		s.Resp.ParamsError(ctx, err)
		return
	}
{{end}}
{{if eq .Method "GET" "DELETE" }}
	if err := ctx.ShouldBindQuery(&in); err != nil {
		s.Resp.ParamsError(ctx, err)
		return
	}
{{else if eq .Method "POST" "PUT" }}
	if err := ctx.ShouldBindJSON(&in); err != nil {
		s.Resp.ParamsError(ctx, err)
		return
	}
{{else}}
	if err := ctx.ShouldBind(&in); err != nil {
		s.Resp.ParamsError(ctx, err)
		return
	}
{{end}}
	md := metadata.New(nil)
	for k, v := range ctx.Request.Header {
		md.Set(k, v...)
	}
	newCtx := metadata.NewIncomingContext(ctx, md)
	out, err := s.Server.({{ $.InterfaceName }}).{{.Name}}(newCtx, &in)
	if err != nil {
		s.Resp.Error(ctx, err)
		return
	}

	s.Resp.Success(ctx, out)
}
{{end}}
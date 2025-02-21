package main

import (
	"flag"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "v1.0.0"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-yiigo %v\n", version)
		return
	}

	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}

const (
	contextPkg = protogen.GoImportPath("context")
	httpPkg    = protogen.GoImportPath("net/http")
	chiPkg     = protogen.GoImportPath("github.com/go-chi/chi/v5")
	contribPkg = protogen.GoImportPath("github.com/yiigo/contrib")
	resultPkg  = protogen.GoImportPath("github.com/yiigo/contrib/result")
)

// generateFile generates a _http.pb.go file containing HTTP service definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_http.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-yiigo. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-yiigo ", version)
	g.P("// - protoc           ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(file, g)
	return g
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

// generateFileContent generates the HTTP service definitions, excluding the package statement.
func generateFileContent(file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}
	for _, service := range file.Services {
		genService(g, service)
	}
}

// Method(context.Context, *MethodReq) (*MethodResp, error)
func serverSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	var reqArgs []string
	// params
	reqArgs = append(reqArgs, "ctx "+g.QualifiedGoIdent(contextPkg.Ident("Context")))
	reqArgs = append(reqArgs, "req *"+g.QualifiedGoIdent(method.Input.GoIdent))
	// return
	resp := "(*" + g.QualifiedGoIdent(method.Output.GoIdent) + ", error)"
	return method.GoName + "(" + strings.Join(reqArgs, ", ") + ") " + resp
}

func genService(g *protogen.GeneratedFile, service *protogen.Service) {
	// Server interface.
	serverType := "Http" + service.GoName
	g.P("// ", serverType, " is the API definition for ", service.GoName)
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
	}
	g.AnnotateSymbol(serverType, protogen.Annotation{Location: service.Location})
	// type HttpXXX interface {
	g.P("type ", serverType, " interface {")
	for _, m := range service.Methods {
		if m.Desc.IsStreamingClient() || m.Desc.IsStreamingServer() {
			continue
		}
		g.AnnotateSymbol(serverType+"."+m.GoName, protogen.Annotation{Location: m.Location})
		if m.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
		}
		g.P(m.Comments.Leading, serverSignature(g, m))
	}
	g.P("}")
	g.P()
	// Register service HttpServer.
	g.P("func Register", serverType, "(r ", chiPkg.Ident("Router"), ", srv ", serverType, ") {")
	for _, m := range service.Methods {
		rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
		if rule != nil && ok {
			method, path := getHttpRouter(rule)
			g.P(strings.TrimSuffix(m.Comments.Leading.String(), "\n"))
			g.P("r.", method, `("`, path, `", _`, service.GoName, "_", m.GoName, `(srv))`)
			// additional bindings
			for _, bind := range rule.AdditionalBindings {
				method, path := getHttpRouter(bind)
				g.P("r.", method, `("`, path, `", _`, service.GoName, "_", m.GoName, `(srv))`)
			}
		}
	}
	g.P("}")
	g.P()
	// Register service methods.
	for _, m := range service.Methods {
		g.P("func _", service.GoName, "_", m.GoName, "(srv ", serverType, ") http.HandlerFunc {")
		g.P("return func(w ", httpPkg.Ident("ResponseWriter"), ", r *", httpPkg.Ident("Request"), ") {")
		g.P("ctx := r.Context()")
		g.P("// parse request")
		g.P("req := new(", m.Input.GoIdent, ")")
		g.P("if err := ", contribPkg.Ident("BindProto"), "(r, req); err != nil {")
		g.P(resultPkg.Ident("Err"), `(err).JSON(w, r)`)
		g.P("return")
		g.P("}")
		g.P("// call service")
		g.P("resp, err := srv.", m.GoName, "(ctx, req)")
		g.P("if err != nil {")
		g.P(resultPkg.Ident("Err"), "(err).JSON(w, r)")
		g.P("return")
		g.P("}")
		g.P(resultPkg.Ident("OK"), "(resp).JSON(w, r)")
		g.P("}")
		g.P("}")
	}
}

func getHttpRouter(rule *annotations.HttpRule) (string, string) {
	switch v := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		return "Get", v.Get
	case *annotations.HttpRule_Put:
		return "Put", v.Put
	case *annotations.HttpRule_Post:
		return "Post", v.Post
	case *annotations.HttpRule_Delete:
		return "Delete", v.Delete
	case *annotations.HttpRule_Patch:
		return "Patch", v.Patch
	case *annotations.HttpRule_Custom:
		return v.Custom.GetKind(), v.Custom.GetPath()
	}
	return "Unknown", ""
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }

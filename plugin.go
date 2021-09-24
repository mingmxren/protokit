package protokit

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"io/ioutil"
	"os"
)

// Plugin describes an interface for running protoc code generator plugins
type Plugin interface {
	Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)
}

// RunPlugin runs the supplied protokit by reading input from stdin and generating output to stdout.
func RunPlugin(p Plugin) error {
	return RunPluginWithIO(p, os.Stdin, os.Stdout)
}

// RunPluginWithIO runs the supplied protokit using the supplied reader and writer for IO.
func RunPluginWithIO(p Plugin, r io.Reader, w io.Writer) error {
	req, err := readRequest(r)
	if err != nil {
		return err
	}

	resp, err := p.Generate(req)
	if err != nil {
		return err
	}

	return writeResponse(w, resp)
}

func readRequest(r io.Reader) (*pluginpb.CodeGeneratorRequest, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	req := new(pluginpb.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		return nil, err
	}

	if len(req.GetFileToGenerate()) == 0 {
		return nil, fmt.Errorf("no files were supplied to the generator")
	}

	return req, nil
}

func writeResponse(w io.Writer, resp *pluginpb.CodeGeneratorResponse) error {
	data, err := proto.Marshal(resp)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

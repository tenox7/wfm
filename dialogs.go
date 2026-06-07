package main

import (
	"bytes"
	"fmt"
	"html"
	"runtime"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/afero"
)

type promptPage struct {
	chrome
	Action   string
	FileName string
	Detail   string
	Options  string
	Items    []string
	RW       bool
}

type editPage struct {
	chrome
	FileName string
	Options  string
	Content  string
	RW       bool
}

type aboutPage struct {
	chrome
	Vers    string
	Runtime string
}

func selOpt(s string, f ...struct{ v, n string }) string {
	var o []string
	var m = make(map[string]string)
	m[s] = "SELECTED"
	m[""] = "DISABLED"
	for _, i := range f {
		o = append(o, fmt.Sprintf("<OPTION VALUE=\"%v\" %v>%v</OPTION>", i.v, m[i.v], i.n))
	}
	return strings.Join(o, "\n")
}

func escapeAll(s []string) []string {
	o := make([]string, len(s))
	for i, v := range s {
		o[i] = html.EscapeString(v)
	}
	return o
}

func (r *wfmRequest) prompt(action string, mul []string) {
	data := promptPage{
		chrome: r.chrome(""),
		Action: action,
		RW:     r.rwAccess,
	}

	switch action {
	case "rename":
		data.FileName = html.EscapeString(r.uFbn)
	case "move":
		data.FileName = html.EscapeString(r.uFbn)
		data.Options = upDnDir(r.uDir, "", r.fs)
	case "delete":
		fi, _ := r.fs.Stat(r.uDir + "/" + r.uFbn)
		if fi.IsDir() {
			data.Detail = "directory - recursively"
		} else {
			data.Detail = "file, size " + humanize.Bytes(uint64(fi.Size()))
		}
		data.FileName = html.EscapeString(r.uFbn)
	case "multi_delete":
		data.Items = escapeAll(mul)
	case "multi_move":
		data.Options = upDnDir(r.uDir, r.uFbn, r.fs)
		data.Items = escapeAll(mul)
	}

	r.render(action, data)
}

func (r *wfmRequest) editText() {
	fi, err := r.fs.Stat(r.uDir + "/" + r.uFbn)
	if err != nil {
		htErr(r.w, "Unable to get file attributes", err)
		return
	}
	if fi.Size() > 1<<20 {
		htErr(r.w, "edit", fmt.Errorf("the file is too large for editing"))
		return
	}
	f, err := afero.ReadFile(r.fs, r.uDir+"/"+r.uFbn)
	if err != nil {
		htErr(r.w, "Unable to read file", err)
		return
	}
	le := *defLe
	if bytes.IndexByte(f, '\r') != -1 {
		le = "CRLF"
	}
	css := ""
	if r.modern {
		css = `html,body{height:100%;margin:0}form{height:100%;display:flex;flex-direction:column}form>textarea{flex:1 1 0;min-height:0;width:100%;box-sizing:border-box;resize:none}`
	}
	r.render("edit", editPage{
		chrome:   r.chrome(css),
		FileName: html.EscapeString(r.uFbn),
		Options: selOpt(le, []struct{ v, n string }{
			{"LF", "LF (Unix)"},
			{"CRLF", "CRLF (Windows)"},
		}...),
		Content: html.EscapeString(string(f)),
		RW:      r.rwAccess,
	})
}

func (r *wfmRequest) about(ua string) {
	r.render("about", aboutPage{
		chrome: r.chrome(""),
		Vers:   vers,
		Runtime: fmt.Sprintf("Build: %v %v-%v<BR>Agent: %v<P>",
			runtime.Version(),
			runtime.GOARCH,
			runtime.GOOS,
			html.EscapeString(ua)),
	})
}

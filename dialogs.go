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

// CodeMirror assets loaded from the CDN, relative to -codemirror_url (npm
// package layout). Loaded only for -textedit=codemirror in modern browsers.
var (
	cmStyles = []string{
		"lib/codemirror.min.css",
		"addon/dialog/dialog.min.css",             // search box
		"addon/fold/foldgutter.min.css",           // fold markers
		"addon/display/fullscreen.min.css",        // F11 fullscreen
		"addon/search/matchesonscrollbar.min.css", // search hits on scrollbar
	}
	// Order matters: core first, then addons; an addon that uses another must
	// come after it (xml-fold before matchtags/closetag, dialog+searchcursor
	// before search, foldcode before foldgutter).
	cmScripts = []string{
		"lib/codemirror.min.js", // core, must be first
		"mode/meta.min.js",      // findModeByFileName for syntax detection
		"addon/mode/loadmode.min.js",
		// find / replace / jump-to-line (Ctrl-F, Ctrl-G, Shift-Ctrl-F/R, Alt-G)
		"addon/dialog/dialog.min.js",
		"addon/search/searchcursor.min.js",
		"addon/scroll/annotatescrollbar.min.js",
		"addon/search/matchesonscrollbar.min.js",
		"addon/search/search.min.js",
		"addon/search/jump-to-line.min.js",
		// bracket / tag matching and auto-closing
		"addon/edit/matchbrackets.min.js",
		"addon/edit/closebrackets.min.js",
		"addon/fold/xml-fold.min.js", // needed by matchtags, closetag and folding
		"addon/edit/matchtags.min.js",
		"addon/edit/closetag.min.js",
		// editing aids
		"addon/edit/trailingspace.min.js",
		"addon/comment/comment.min.js", // Ctrl-/ toggle comment
		"addon/selection/active-line.min.js",
		// code folding
		"addon/fold/foldcode.min.js",
		"addon/fold/foldgutter.min.js",
		"addon/fold/brace-fold.min.js",
		"addon/fold/indent-fold.min.js",
		"addon/fold/comment-fold.min.js",
		// fullscreen toggle
		"addon/display/fullscreen.min.js",
	}
)

type promptPage struct {
	chrome
	Action   string // human label + fallback fn dispatch value
	Op       string // op dispatch value for single-file dialogs (re/mv/rm)
	FileName string
	Detail   string
	Options  string
	Items    []string
	RW       bool
}

type editPage struct {
	chrome
	FileName   string
	Options    string
	Content    string
	RW         bool
	CodeMirror bool
	CMBase     string
	CMStyles   []string
	CMScripts  []string
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

	// single-file dialogs commit via POST to the file's own path (?op=CODE);
	// dir-context dialogs (mkdir/multi_*) keep posting to the prefix with fn.
	switch action {
	case "rename":
		data.Op = "re"
		data.FormAction = html.EscapeString(r.pathURLRaw(r.uDir, r.uFbn, nil))
		data.FileName = html.EscapeString(r.uFbn)
	case "move":
		data.Op = "mv"
		data.FormAction = html.EscapeString(r.pathURLRaw(r.uDir, r.uFbn, nil))
		data.FileName = html.EscapeString(r.uFbn)
		data.Options = upDnDir(r.uDir, "", r.fs)
	case "delete":
		data.Op = "rm"
		data.FormAction = html.EscapeString(r.pathURLRaw(r.uDir, r.uFbn, nil))
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
		r.htErr("Unable to get file attributes", err)
		return
	}
	if fi.Size() > 1<<20 {
		r.htErr("edit", fmt.Errorf("the file is too large for editing"))
		return
	}
	f, err := afero.ReadFile(r.fs, r.uDir+"/"+r.uFbn)
	if err != nil {
		r.htErr("Unable to read file", err)
		return
	}
	le := *defLe
	if bytes.IndexByte(f, '\r') != -1 {
		le = "CRLF"
	}
	// CodeMirror is opt-in (--textedit=codemirror) and modern-mode only; legacy
	// browsers always get a plain textarea.
	cm := r.modern && *textEdit == "codemirror"
	css := ""
	if r.modern {
		// Make the textarea fill the viewport below the toolbar.
		css = `html,body{height:100%;margin:0}form{height:100%;display:flex;flex-direction:column}form>textarea{flex:1 1 0;min-height:0;width:100%;box-sizing:border-box;resize:none}`
		if cm {
			// CodeMirror hides the textarea and inserts a .CodeMirror sibling;
			// size it the same way. The textarea rule above stays as the
			// fallback layout if the CDN never loads. The last rule tints
			// trailing whitespace highlighted by the trailingspace addon.
			css += `form>.CodeMirror{flex:1 1 0;min-height:0;height:auto}.cm-trailingspace{background-color:#FFE0E0}`
		}
	} else {
		css = `html,body{height:100%}form{height:100%}textarea{width:98%;height:80%}`
	}
	r.render("edit", editPage{
		chrome:   r.chrome(css),
		FileName: html.EscapeString(r.uFbn),
		Options: selOpt(le, []struct{ v, n string }{
			{"LF", "LF (Unix)"},
			{"CRLF", "CRLF (Windows)"},
		}...),
		Content:    html.EscapeString(string(f)),
		RW:         r.rwAccess,
		CodeMirror: cm,
		CMBase:     strings.TrimRight(*cmCDN, "/"),
		CMStyles:   cmStyles,
		CMScripts:  cmScripts,
	})
}

func (r *wfmRequest) about(ua string) {
	htmlMode := "legacy"
	if r.modern {
		htmlMode = "modern"
	}
	r.render("about", aboutPage{
		chrome: r.chrome(""),
		Vers:   vers,
		Runtime: fmt.Sprintf("Build: %v %v-%v<BR>Agent: %v<BR>HTML: %v<P>",
			runtime.Version(),
			runtime.GOARCH,
			runtime.GOOS,
			html.EscapeString(ua),
			htmlMode),
	})
}

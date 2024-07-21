package game

// HelpPage contains the text of a help page from a mod.
type HelpPage struct {
	Path     string   // Full path to help page with file extension.
	Title    string   // Title of the help page, taken from the first line of the file.
	Contents []string // Contents of the help page minus the title line.
}

// Help pages
var HelpPages = map[string]*HelpPage{}

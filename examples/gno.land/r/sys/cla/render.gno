package cla

import (
	"gno.land/p/moul/helplink"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/ufmt/v0"
)

func Render(path string) string {
	out := md.H1("Contributor License Agreement (CLA)")

	out += md.Paragraph("A Contributor License Agreement (CLA) must be signed before deploying packages.")
	out += md.Paragraph("It grants the necessary rights for your code to be used on-chain.")
	out += md.Paragraph("The CLA document is defined through a " + md.Link("GovDAO", "/r/gov/dao") + " governance proposal.")

	if requiredHash == "" {
		out += md.HorizontalRule()
		out += md.H2("Status")
		out += md.Paragraph(md.Bold("CLA enforcement is currently DISABLED."))
		out += md.Paragraph("All package deployments are allowed.")
		return out
	}

	if claURL != "" {
		out += md.Paragraph("You can read the full agreement here: " + claURL)
	}

	out += md.HorizontalRule()
	out += md.H2("Status")
	out += md.Paragraph(md.Bold("CLA enforcement is ENABLED"))

	table := mdtable.Table{Headers: []string{"", ""}}
	table.Append([]string{md.Bold("Required Hash"), md.InlineCode(requiredHash)})
	table.Append([]string{md.Bold("Signers"), ufmt.Sprintf("%d contributor(s)", signatures.Size())})
	out += table.String()

	out += md.HorizontalRule()
	out += md.H2("How to Sign")
	out += md.Paragraph(md.Bold("From gnoweb:") + " Click the \"Sign CLA\" action link below.")
	out += md.Paragraph(md.Bold("From gnokey:") + " Visit the " + md.Link("helper page", "$help") + " to get the exact command to sign.")

	out += md.H3("Actions")
	out += helplink.Func("Sign CLA", "Sign", "hash", requiredHash) + "\n"
	return out
}

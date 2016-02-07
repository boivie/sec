package cmd
import (
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/storage"
	"github.com/fatih/color"
	"encoding/json"
	"bytes"
)

var CmdDump = cli.Command{
	Name:  "dump",
	Usage: "dump database",
	Action: cmdDump,
}

func cmdDump(c *cli.Context) {
	stor, err := storage.New()
	if err != nil {
		panic("Failed to open storage")
	}
	topic, err := storage.DecodeTopic(c.Args().First())
	if err != nil {
		panic(err)
	}

	header := color.New(color.FgYellow)
	protected := color.New(color.FgCyan)
	payload := color.New(color.FgWhite)

	records, err := stor.GetAll(topic)
	for _, record := range records {
		header.Printf("record %s:%d (%s)\n", topic.Base58(), record.Index, record.Type)
		protected.Printf("protected %s\n", record.Message.Protected)

		var f interface{}
		d := json.NewDecoder(bytes.NewBuffer(record.Message.Payload))
		d.UseNumber()
		if err := d.Decode(&f); err != nil {
			panic(err)
		}
		s, _ := json.MarshalIndent(f, "", "  ")
		payload.Printf("\n%s\n\n", s)
	}
}
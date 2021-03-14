package commands

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

var (
	s3lowaccessCmd = &cobra.Command{
		Use:   "s3lowaccess",
		Run:   s3lowaccessCommand,
		Short: "low level access to AWS S3",
		Long:  "low level access to AWS S3",
	}
)

func s3lowaccessCommand(cmd *cobra.Command, args []string) {
	if err := s3lowaccessAction(args); err != nil {
		Exit(err, 1)
	}
}

func s3lowaccessAction(args []string) error {
	if len(args) == 4 && args[0] == "list" {
		// arciv s3lowaccess list <region> <bucket> <key-prefix>
		blobNames, err := s3Op.findFilePaths(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		for _, blobName := range blobNames {
			fmt.Println(blobName)
		}
		return nil
	}
	if len(args) == 4 && args[0] == "load" {
		// arciv s3lowaccess load <region> <bucket> <key>
		lines, err := s3Op.loadLines(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	}
	if len(args) == 5 && args[0] == "download" {
		// arciv s3lowaccess download <region> <bucket> <key> <write-path> # to deep archive
		return s3Op.receiveBlobs(args[1], args[2], []string{args[4]}, []string{args[3]})
	}
	if len(args) == 5 && args[0] == "upload" {
		// arciv s3lowaccess upload <region> <bucket> <key> <read-path>
		return s3Op.sendBlobs(args[1], args[2], []string{args[4]}, []string{args[3]})
	}
	if len(args) == 4 && args[0] == "write" {
		// arciv s3lowaccess write <region> <bucket> <key> # read from stdin
		stdin := bufio.NewScanner(os.Stdin)
		var lines []string
		for stdin.Scan() {
			lines = append(lines, stdin.Text())
		}
		return s3Op.writeLines(args[1], args[2], args[3], lines)
	}
	if len(args) == 5 && args[0] == "restore" {
		// arciv s3lowaccess restore <resgion> <bucket> <key> <valid-days>
		validDays64, err := strconv.ParseInt(args[4], 10, 32)
		if err != nil {
			return err
		}
		_, err = s3Op.receiveBlobsRequest(args[1], args[2], []string{args[3]}, int32(validDays64))
		return err
	}
	fmt.Println("Usage:")
	fmt.Println("  arciv s3lowaccess list <region> <bucket> <key-prefix>")
	fmt.Println("  arciv s3lowaccess load <region> <bucket> <key> # write to stdout")
	fmt.Println("  arciv s3lowaccess download <region> <bucket> <key> <write-path>")
	fmt.Println("  arciv s3lowaccess upload <region> <bucket> <key> <read-path> # to deep archive")
	fmt.Println("  arciv s3lowaccess write <region> <bucket> <key> # read from stdin")
	fmt.Println("  arciv s3lowaccess restore <region> <bucket> <key> <valid-days> # restore from deep archive")
	return nil
}

func init() {
	RootCmd.AddCommand(s3lowaccessCmd)
}

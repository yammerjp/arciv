package commands

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
	bucketName = "arciv-development-backet"
	if len(args) == 1 && args[0] == "list-blob" {
		// arciv s3lowaccess list-blob
		prepareS3BucketClient()
		blobNames, err := s3Op.listBlobs()
		if err != nil {
			return err
		}
		for _, blobName := range blobNames {
			fmt.Println(blobName)
		}
		return nil
	}
	if len(args) == 2 && args[0] == "load" {
		// arciv s3lowaccess load <key>
		prepareS3BucketClient()
		lines, err := s3Op.loadLines(args[1])
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	}
	if len(args) == 3 && args[0] == "download" {
		// arciv s3lowaccess download <key> <write-path>
		prepareS3BucketClient()
		return s3Op.receiveBlobs([]string{args[2]}, []string{args[1]})
	}
	if len(args) == 3 && args[0] == "upload" {
		// arciv s3lowaccess upload <key> <read-path>
		prepareS3BucketClient()
		return s3Op.sendBlobs([]string{args[2]}, []string{args[1]})
	}
	if len(args) == 2 && args[0] == "write" {
		// arciv s3lowaccess write <key> (read from stdin)
		stdin := bufio.NewScanner(os.Stdin)
		var lines []string
		for stdin.Scan() {
			lines = append(lines, stdin.Text())
		}
		prepareS3BucketClient()
		return s3Op.writeLines(args[1], lines)
	}
	fmt.Println("Usage:")
	fmt.Println("  arciv s3lowaccess list-blob")
	fmt.Println("  arciv s3lowaccess load <key>")
	fmt.Println("  arciv s3lowaccess download <key> <write-path>")
	fmt.Println("  arciv s3lowaccess upload <key> <read-path>")
	fmt.Println("  arciv s3lowaccess write <key> (read from stdin)")
	return nil
}

func init() {
	RootCmd.AddCommand(s3lowaccessCmd)
}

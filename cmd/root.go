/*
Copyright © 2022 Giovanni Orciuolo <giovanni@orciuolo.it>
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "facebook-video-downloader",
	Short: "Download a Facebook video by providing an URL",
	Long: `Usage example:

facebook-video-downloader <video_url>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return nil
		}

		uri := args[0]
		outPath := cmd.Flag("out").Value.String()
		if outPath == "" {
			fmt.Println("Please specify video output path with --out or -o")
			return nil
		}

		res, err := http.Get(strings.Replace(uri, "www", "mbasic", 1))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			log.Fatalf("response not ok: %d %s", res.StatusCode, res.Status)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return err
		}

		var videoHref string

		doc.Find("a[target='_blank']").EachWithBreak(func(i int, s *goquery.Selection) bool {
			href, found := s.Attr("href")
			if !found || !strings.Contains(href, "video_redirect") {
				return true // meaning continue
			}

			href, err := url.QueryUnescape(href[strings.Index(href, "=")+1 : strings.Index(href, "&")])
			if err != nil {
				fmt.Println("error unescaping href: " + err.Error())
				return true
			}

			videoHref = href
			return false // found, break out of loop
		})

		if videoHref == "" {
			return errors.New("no video found in page")
		}

		videoRes, err := http.Get(videoHref)
		if err != nil {
			return err
		}
		defer videoRes.Body.Close()

		videoBytes, err := ioutil.ReadAll(videoRes.Body)
		if err != nil {
			return err
		}

		if outPath != "" {
			err = ioutil.WriteFile(outPath, videoBytes, fs.ModeAppend)
			return err
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("out", "o", "", "video output path")
}

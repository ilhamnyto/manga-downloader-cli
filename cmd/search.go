/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/manifoldco/promptui"
	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search your favorite manga.",
	Long: `Search your favorite manga and choose which one to download.`,
	Run: func(cmd *cobra.Command, args []string) {
		downloadManga()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type promptContent struct {
	errorMsg 	string
	label		string
}

type MangaDetail struct {
	Title	string 		`json:"title"`
	URL		string		`json:"url"`
}

func SearchManga(keyword string, c *colly.Collector) []MangaDetail {

	mangaList := make([]MangaDetail, 0, 1)
	c.OnHTML("div.list-update_item", func(e *colly.HTMLElement) {
		manga := MangaDetail{Title: e.ChildText("h3.title"), URL: e.ChildAttr("a.data-tooltip", "href")}
		mangaList = append(mangaList, manga)
	})

	c.Visit("https://komikcast.io/?s=" + strings.ReplaceAll(keyword, " ", "+"))
	return mangaList
}

func GetMangaChapter(s *MangaDetail, c *colly.Collector) []MangaDetail {
	mc := make([]MangaDetail, 0, 1)
	c.OnHTML("li.komik_info-chapters-item", func(e *colly.HTMLElement) {
		chapter := MangaDetail{Title: e.ChildText("a.chapter-link-item"), URL: e.ChildAttr("a.chapter-link-item", "href")}

		mc = append(mc, chapter)
	})

	c.Visit(s.URL)
	
	return mc
}

func GetPDF(mc *MangaDetail, c *colly.Collector, filename string) {
	imgUrl := make([]string, 0, 1)

	c.OnHTML("div.main-reading-area > img", func(e *colly.HTMLElement) {
		img := e.Attr("src")
		imgUrl = append(imgUrl, img)
	})

	c.Visit(mc.URL)

	pdf := gopdf.GoPdf{}

	pdf.Start(gopdf.Config{})

	for _, url := range(imgUrl) {

		resp, err := http.Get(url)

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		
		imageData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		
		img, _, err := image.Decode(bytes.NewReader(imageData))
		
		if err != nil {
			panic(err)
		}

		pdf.AddPageWithOption(gopdf.PageOption{PageSize: &gopdf.Rect{W: float64(img.Bounds().Dx()), H: float64(img.Bounds().Dy())}})
		
		ii, err := gopdf.ImageHolderByBytes(imageData)

		if err != nil {
			panic(err)
		}
		
		pdf.ImageByHolder(ii, 0, 0, &gopdf.Rect{W: float64(img.Bounds().Dx()), H: float64(img.Bounds().Dy())})

	}

	err := pdf.WritePdf(filename+".pdf")

	if err != nil {
		panic(err)
	}
}

func promptGetInput(pc promptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(pc.errorMsg)
		}

		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt: "{{ . }}",
		Valid: "{{ . | green }}",
		Invalid: "{{ . | red }}",
		Success: "{{ . | bold}}",
	}

	prompt := promptui.Prompt{
		Label: pc.label,
		Templates: templates,
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("-> %s\n", result)

	return result
}

func promptGetSelect(pc promptContent, data []MangaDetail) int {
	items := make([]string, 0, 1)
	index := -1

	for _, name := range(data) {
		items = append(items, name.Title)
	}

	var (
		result string
		err error
	)

	for index < 0 {
		prompt := promptui.Select{
			Label: pc.label,
			Items: items,
		}

		index, result, err = prompt.Run()
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("-> %s\n", result)

	return index
}

func downloadManga() {
	c := colly.NewCollector(
		colly.AllowedDomains("komikcast.io", "www.komikcast.io"),
	)

	searchMangaPromptContent := promptContent{
		errorMsg: "Please provide the manga name.",
		label: "What manga would you like to search? ",
	}

	keyword := promptGetInput(searchMangaPromptContent)

	mangaList := SearchManga(keyword, c)

	selectMangaPromptContent := promptContent{
		errorMsg: "Please select the manga.",
		label: "Which manga would you like to download? ",
	}

	mangaIndex := promptGetSelect(selectMangaPromptContent, mangaList)

	chapterList := GetMangaChapter(&mangaList[mangaIndex], c)

	selectChapterPromptContent := promptContent{
		errorMsg: "Please select the chapter.",
		label: "Which chapter would you like to download? ",
	}

	chapterIndex := promptGetSelect(selectChapterPromptContent, chapterList)
	filename := mangaList[mangaIndex].Title + " " + chapterList[chapterIndex].Title
	GetPDF(&chapterList[chapterIndex], c, filename)
}
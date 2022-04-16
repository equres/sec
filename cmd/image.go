// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"fmt"
	"image/color"
	"os"

	"github.com/fogleman/gg"
	"github.com/spf13/cobra"
)

var ImagePath string
var Title string

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Generate an image to use in Social Media posts",
	Long:  `Generate an image to use in Social Media posts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stat(ImagePath)
		if err != nil {
			return fmt.Errorf("No image exists on this path: %v", ImagePath)
		}

		bgImage, err := gg.LoadImage(ImagePath)
		if err != nil {
			return err
		}

		dc := gg.NewContext(1200, 628)

		dc.DrawImage(bgImage, 0, 0)
		margin := 20.0
		x := margin
		y := margin

		w := float64(dc.Width()) - (2.0 * margin)
		h := float64(dc.Height()) - (2.0 * margin)
		dc.SetColor(color.RGBA{0, 0, 0, 204})
		dc.DrawRectangle(x, y, w, h)
		dc.Fill()

		err = addCompanyName(dc)
		if err != nil {
			return err
		}

		err = addCompanyLink(dc)
		if err != nil {
			return err
		}

		err = addTitle(dc, Title)
		if err != nil {
			return err
		}

		err = dc.SavePNG("out.png")
		if err != nil {
			return err
		}

		return nil
	},
}

func addCompanyName(dc *gg.Context) error {
	err := dc.LoadFontFace("./static/OpenSans-Bold.ttf", 50)
	if err != nil {
		return err
	}
	dc.SetColor(color.White)
	marginX := 50.0
	marginY := 30.0
	textWidth, textHeight := dc.MeasureString("Equres LLC")
	x := float64(dc.Width()) - textWidth - marginX
	y := float64(dc.Height()) - textHeight - marginY
	dc.DrawString("Equres LLC", x, y)
	return nil
}

func addCompanyLink(dc *gg.Context) error {
	err := dc.LoadFontFace("./static/OpenSans-Light.ttf", 50)
	if err != nil {
		return err
	}
	textColor := color.White
	r, g, b, _ := textColor.RGBA()
	mutedColor := color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(200),
	}
	dc.SetColor(mutedColor)
	_, textHeight := dc.MeasureString("https://equres.com")
	x := 70.0
	y := float64(dc.Height()) - textHeight - 40.0
	dc.DrawString("https://equres.com", x, y)
	return nil
}

func addTitle(dc *gg.Context, title string) error {
	if err := dc.LoadFontFace("./static/OpenSans-Bold.ttf", 75); err != nil {
		return err
	}
	textRightMargin := 60.0
	textTopMargin := 90.0
	x := textRightMargin
	y := textTopMargin
	maxWidth := float64(dc.Width()) - textRightMargin - textRightMargin
	dc.SetColor(color.Black)
	dc.DrawStringWrapped(title, x+1, y+1, 0, 0, maxWidth, 1.5, gg.AlignLeft)
	dc.SetColor(color.White)
	dc.DrawStringWrapped(title, x, y, 0, 0, maxWidth, 1.5, gg.AlignLeft)
	return nil
}

func init() {
	rootCmd.AddCommand(imageCmd)

	imageCmd.PersistentFlags().StringVarP(&ImagePath, "image-path", "i", "./static/image.png", "define the image path")
	imageCmd.PersistentFlags().StringVarP(&Title, "title", "t", "", "set the title to be written in the image")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// imageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// imageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

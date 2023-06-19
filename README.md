# REST API with Clean Architecture

This project is a personal learning project aimed at building a CLI to download mangas into PDF.

## Features

- Search manga
- List manga chapters
- Download manga into pdf

## Installation

To run this project locally, follow these steps:

1. Clone the repository: `git clone https://github.com/ilhamnyto/manga-downloader-cli.git`
2. Install the required dependencies: `go mod download`
3. Build the app: `go run build .`
4. Run app `./manga-downloader-cli search`

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgments

This project was made possible by the following open-source libraries:

- [cobra](https://github.com/spf13/cobra)
- [promptui](https://github.com/manifoldco/promptui)
- [colly](https://github.com/gocolly/colly)
- [gopdf](github.com/signintech/gopdf)
# crb2pdf

Use to convert CRB files to PDF

## Usage
```
cbr2pdf - an utility for converting CBR to PDF. Default resolution is 1072x1448 (Pocketbook Touch HD 3)

Usage: cbr2pdf <source file> [destination file]

Environment:    WIDTH  ... specify X resolution of your reader
                HEIGHT ... specify Y resolution of your reader

Examples: ./cbr2pdf my-favorite-comicbook.cbr output.pdf

          WIDTH=758 HEIGHT=1024 ./cbr2pdf my-favorite-comicbook.cbr output.pdf       # Pocketbook Touch Lux 4
```

## Build
```
go build
```
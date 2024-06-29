# Annotate

A simple text-based tool for annotating documents in the command line.

## Usage

To annotate a CSV file called `to_annotate.csv`, with a column called `text-column-name` run

```bash
annotate -i to_annotate.csv -o annotated.csv -t text-column-name -a annotation
```

This will open the terminal user interface allowing you to annotate each row of the file.
The resulting annotations will be written as a CSV file to `annotated.csv`, with the title of each item being the first column in the original file, the second column being the text column, and the third column being the manual annotation.


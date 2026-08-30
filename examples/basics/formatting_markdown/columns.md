# Columns

This is some standard text without further formatting.
It uses the whole line.
But the following text will consist of 2/3 text on the left side vs 1/3 table on the right side.

<!--
    Specifies that in the following sectio, columns are being used.
    The number of columns is dynamically adjusted to how you use the ".col-X" modifier (up to .col-4).
    So in the example below, the card is divided into 3 columns (2+1), where the first block
    spreads over the first two columns and the last block over the 3rd column.
-->

::: {.columns}
<!-- The following block is 2 elements wide -->
::: {.col-2}
This text uses the first 2/3 of the whole width.
The last 1/3 is used by the table.
:::

<!-- This block is only 1 element wide -->
::: {.col-1}
| A     | B     |
| :---: | :---: |
| 7     | 5     |
:::
<!-- the following line ends the {.columns} -->
:::

---

````markdown
This is some standard text without further
formatting. It uses the whole line.
But the following text will consist of
2/3 text on the left side vs 1/3 table on
the right side.

::: {.columns}
<!-- The following block is 2 elements wide -->
::: {.col-2}
This text uses the first 2/3 of the whole width.
The last 1/3 is used by the table.
:::
<!--  This block is only 1 element wide -->
::: {.col-1}
| A     | B     |
| :---: | :---: |
| 7     | 5     |
| -123  | 14    |
:::
<!-- The following line ends the {.columns} -->
:::
````

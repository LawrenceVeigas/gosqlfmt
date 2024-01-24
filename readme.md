# gosqlfmt

This project is being developed to format databricks flavoured queries.

### To do

- assign random string to contents of each bracket
- write function to parse contents of bracket and categorize query as
  - create or replace AS
  - create or replace (
  - insert
  - merge
  - select :check
  - delete
  - set
- post categorization of query, run respective format function to format the contents of the bracket
- once formatted, replace random string with new formatted text

# gosqlfmt

This project is being developed to format databricks flavoured queries.

### To do

- algorithm to replace brackets in correct order (for heavily nested queries)
- assign random string to contents of each bracket *Done*
- write function to parse contents of bracket and categorize query as
  - create or replace AS
  - create or replace (
  - insert
  - merge
  - select *Done*
  - delete
  - set *Done*
  - union (set operations)
- post categorization of query, run respective format function to format the contents of the bracket *Done*
- once formatted, replace random string with new formatted text *Done*

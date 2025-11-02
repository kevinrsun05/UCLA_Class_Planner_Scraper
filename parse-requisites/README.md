We connect to the database and lad the subject names from the subject_areas col in the course offerings table.
We look through course requisites_text if the cell is not Null, then we parse it and try to match it to a basic pattern.
If the pattern matches then parse the requisites to populate requisites_parsed
More complex patterns are not getting matched rn, I am working to do that in the future
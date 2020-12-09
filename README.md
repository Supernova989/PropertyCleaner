## Description
The app takes 3 arguments, all of them are required:
* --dir
* --dicts
* --exts

The _dir_ argument should be an absolute path to the folder where the search is going to be made.
The _dicts_ and _exts_ arguments accept paths separated by comma.
The _dicts_ argument is for paths for property dictionaries (.properties)
The _exts_ argument is a sequence of file extensions. Files with such extensions will be covered in the search.

## Important
Every time when the application starts, it erases the folder "build" located next to the app file.

To compile an app: ```go build -o cleaner```

## Example
```./cleaner --dir=/Users/johndoe/WEB/project_folder/src/ --exts=.html,.java,.js  --dicts=/Users/johndoe/WEB/project_folder/src/main/resources/messages.properties,/Users/johndoe/WEB/project_folder/src/main/resources/messages_de.properties```


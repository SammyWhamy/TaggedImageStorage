## To do list
- [ ] **Bug** • Index file expands with 0 bytes, causing the file to grow to more than twice of its necessary size
- [ ] **Feature** • Add support for removing files
- [ ] **Feature** • Add support for editing tags on existing files
- [ ] **Feature** • Add support for getting single random files
- [ ] **Refactor** • Make code more readable and refactor into different files
- [ ] **Docs** • Add comments to explain confusing parts of the code
- [ ] **Testing** • Add automated testing
- [ ] **Feature** • Add support for the rest of the flags and commands that have not been implemented yet

## Steps to run TIS

```shell
go build # Builds the binary

tis init # Creates an empty index file

tis info # Confirm the index file is created and accessible, should return some basic information.

tis add-file "path/to/file" "semicolon;separated;tags" # Adds a file to the index

tis list "semicolon;separated;tags" # Lists all files with the given tags
# Note that this lists all files with each tag. Meaning a file will show up 3 times if
# it has all 3 tags. This could later be changed with a flag or the consumer could filter the results.
```

### Preparing a batch of test data
- Create a folder called `test_images`
- Create a file called `test_images/.TAGLIST`
- Add images to the `test_images` folder
- Populate the `.TAGLIST` file with the following format:
```
path/to/image1.jpg tag1;tag2;tag3
path/to/image2.jpg tag1;tag2;tag3
path/to/image3.jpg tag1;tag2;tag3
```
- Run the `batch_test.bat` file (Windows only)
- Run `tis info` to confirm all files have been added

## Documentation

### Commands
```shell
# Help
tis help / --help / -h

# Version
tis version / --version / -v

# Initialize the index file
tis init

# Query info about the index
tis info

# Add or remove files from the index
tis add-file "path/to/file" "semicolon;separated;tags"
tis remove-file "path/to/file" # Not implemented yet

# Query the index for files
tis list "semicolon;separated;tags"
tis random ["semicolon;separated;tags"]

# Edit existing tags on a file
tis add-tag "path/to/file" "semicolon;separated;tags" # Not implemented yet
tis remove-tag "path/to/file" "semicolon;separated;tags" # Not implemented yet

# List tags for a file
tis file-tags "path/to/file" # Not implemented yet

# Globally remove tags
tis delete-tag "semicolon;separated;tags" # Not implemented yet

# Rename a tag
tis rename-tag "old_tag" "new_tag" # Not implemented yet

# Rename data folder
tis rename-data-folder "new_folder_name" # Not implemented yet

# Export the index file
tis export ["path/to/export/file"] # Not implemented yet
```

### Flags
```shell
# Verbose
tis --verbose, -V

# Specify file name, use * as filename to use a random file name that keeps the extension
tis add-file "path/to/file" "semicolon;separated;tags" --file-name="new_file_name"

# Don't move the file to the data folder
tis add-file "path/to/file" "semicolon;separated;tags" --no-movet

# Mode for listing files
tis list "semicolon;separated;tags" --exlusive
```

### Notes:
- If a file has no tags (after removing a tag), it will be automatically moved to a `no-tags` folder.
- If a tag has no files (after removing files), it will be automatically removed from the index.

### Index file structure
```shell
> REGION BASIC INFO

# File header
0x2E 0x54 0x49 0x53 # 0 .TIS

0x00 0x00 0x00 0x00 # 4 Tag count
0x00 0x00 0x00 0x00 # 8 File count

0x00 0x00 0x00 0x00 # 12 Tag list offset
0x00 0x00 0x00 0x00 # 16 Tag list length

0x00 0x00 0x00 0x00 # 20 Index offset
0x00 0x00 0x00 0x00 # 24 Index length

0x00 0x00 0x00 0x00 # 28 Length of path to data folder
# Path to data folder

> REGION TAG LIST, REPEATED FOR EACH TAG

0x00 0x00 0x00 0x00 # Offset to file list
0x00 0x00 0x00 0x00 # Length of tag name
# Tag name

> REGION FILE LIST, REPEATED FOR EACH FILE LIST

0x00 0x00 0x00 0x00 # Length of file list
# File list
```
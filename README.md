### Steps to run TIS

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
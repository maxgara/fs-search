# fs-search
- very basic file search functionality
- initially no indexing will be used
- program will construct a list of all words in all files and all locations where words can be found
- list will be stored as an array of hashes and file indexes, written to disk
- list will be sorted by hash (search key)
- searches will use binary search to find key in sorted list
- later iterations may experiment with indexing and other optomization strategies


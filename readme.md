# Split a file into 5MB chunks
go run filetool.go split --input large.zip --size 5MB

# Join back into large.zip.joined
go run filetool.go join --input large.zip.part0
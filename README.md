# sort_images

* Parse flags DONE
```
  -destination string
        Folder to move the images into. The folder is created if it does not exist. (Required)
  
  -dry-run bool
        Set to false to actually make changes. (default true)
  
  -source string
        Folder with unorganised images. Must be an existing folder. (default "./")
```
* validate flags DONE

* Traverse the image-folder and find images DONE

* Find the date when the picture is taken and create a list of directories that is needed to organise the images into DONE

* If dry-run=false create the folder structure 

* Move each image into the correct folder DONE

* If the image already exists there, add a number to the image

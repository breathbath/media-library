# Service for storing and serving images over http

[![Travis Status for breathbath/media-library](https://api.travis-ci.org/breathbath/media-library.svg?branch=master&label=linux+build)](https://travis-ci.org/breathbath/media-library)

Media library has purpose to provide and manage in-house images over http(s). Current implementation has following features:
1. Store http(s) posted images.
2. Storing images in unique folders, so different images with the same name will not be conflicting.
3. JWT-token based auth for posting and deleting images.
4. Size and mime-type validation for incoming images.
5. Automatic cropping and quality optimisation for too large images to save space
6. Serving images over http(s)
7. Dynamic images resizing for fixed width and height values or proportional rezising when one dimension is not specified.
8. Caching of resized images
9. Multi-file uploading with multipart form content
10. Proxy support

## Configuration options

The service is configurable with env variables:

### ASSETS_PATH

_Required_

File system path where to store the whole library payload

    ASSETS_PATH=/media/data/images

### HOST

_Required_

Host and port where to run the service:

    HOST=:9295

### TOKEN_ISSUER

_Required_

Jwt token issuer field to differentiate tokens for different environments

    TOKEN_ISSUER=production-media-service

### TOKEN_SECRET

_Required_

Password to deconde/encode the JWT token

    TOKEN_SECRET=dfasdfs

### URL_PREFIX

_Required_

All image urls will be prefix with URL_PREFIX. It's needed for the case where the same domain is used for both media library and other services.

    URL_PREFIX=/media/images/

### MAX_UPLOADED_FILE_MB

_Default 20, float_

Value to limit the max size of one uploaded file. All files over this size will be rejected.

    MAX_UPLOADED_FILE_MB=20

### COMPRESS_JPG_QUALITY

_Default 85, int_

Percent value which will be used to compress all posted jpeg images before saving.

    COMPRESS_JPG_QUALITY=85
    
### VERT_MAX_IMAGE_WIDTH

_Default 0, int_

Maximal width in pixels for images with a vertical orientation (where height > width). 
If a vertical image with a larger width is posted, it will be resized to this width, height will be adjusted proportionally.
The value is needed to spare disk space. Images with big dimensions will be automatically adjusted to the defined size.

    VERT_MAX_IMAGE_WIDTH=800
    
    
### HORIZ_MAX_IMAGE_HEIGHT

_Default 0, int_

Maximal height in pixels for images with a horizontal orientation (where width > height). 
If a horizontal image with a larger height is posted, it will be resized to this height, width will be adjusted proportionally.
The value is needed to spare disk space. Images with big dimensions will be automatically adjusted to the defined size.

    HORIZ_MAX_IMAGE_HEIGHT=600

### PROXY_URL

_Default '', string_

If an image is not found in the local file system, it will be fetched from proxy url. If this option is empty, 404 will be returned.
This option is useful when you have multiple envs but don't want to synchronize whole image set between them. Then you can use prod as proxy to serve images missing in your testing env.

## To start project with docker-compose
    
    docker-compose up -d

## To upload new image
    
    curl -F 'files[]=@/home/me/images/photo1@2x.jpg' -F 'files[]=@/home/me/images/photo2@2x.jpg' -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJtZWRpYV9zZXJ2aWNlIiwiZXhwIjoxNTY3NjMxNDQwLCJpYXQiOjE1NjUwMzk0NDAsImlzcyI6Im1lZGlhLXNlcnZpY2UtZGV2ZWxvcGVyIiwic3ViIjoibWVkaWEtc2VydmVyLWRldiJ9.2K1ueLVk_NrSNgViRl-AmeY-do3WLTFD1We1GiQlwrY' http://localhost:9295/media/images/
    
The response will be similar to this:

    {
        "filepathes": [
            "5d489b785c7a8/photo1_2x.jpg",
            "5d489b785c7a8/photo2_2x.jpg"
        ]
    }
    
## To get file displayed in full size use

    http://localhost:9295/media/images/5d489b785c7a8/photo1_2x.jpg

## To display file cropped

    #200x200
    http://localhost:9295/media/images/200x200/5d489b785c7a8/photo1_2x.jpg
    
    #200x{proportional}
    http://localhost:9295/media/images/200x/5d489b785c7a8/photo1_2x.jpg
    
    #{proportional}x200
    http://localhost:9295/media/images/x200/5d489b785c7a8/photo1_2x.jpg
    
## To generate new token
    
    docker-compose exec media /root/media token media-server-dev

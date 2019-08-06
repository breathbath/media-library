## To upload new image
    
    curl -F 'files=@/Users/breathbath/Desktop/images/photo1@2x.jpg' -F 'files=@/Users/breathbath/Desktop/images/photo2@2x.jpg' -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJtZWRpYV9zZXJ2aWNlIiwiZXhwIjoxNTY3NjMxNDQwLCJpYXQiOjE1NjUwMzk0NDAsImlzcyI6Im1lZGlhLXNlcnZpY2UtZGV2ZWxvcGVyIiwic3ViIjoibWVkaWEtc2VydmVyLWRldiJ9.2K1ueLVk_NrSNgViRl-AmeY-do3WLTFD1We1GiQlwrY' http://localhost:9295/media/images/
    
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
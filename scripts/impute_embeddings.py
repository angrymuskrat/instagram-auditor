from pymongo import MongoClient
import base64
import io
from PIL import Image
import sys
sys.path.insert(0, '../img2vec/')
from img2vec import Image2Vec
import threading


username = 'username'
fullname = 'fullname'
biography = 'biography'
followed_by = 'followedby'
follow = 'follow'
is_business_account = 'isbusinessaccount'
is_joined_recently = 'isjoinedrecently'
business_category_name = 'businesscategoryname'
business_category_name = 'categoryid'
is_private = 'isprivate'
is_verified = 'isverified'
has_connected_fb = 'connectedfbpage'
profile_pic = 'profilepic'
profile_pic_url = 'profilepicurl'
posts_count = 'postscount'
posts = 'posts'
image = 'image'
profile_pic_embedding = 'profilepicembedding'
image_embedding = 'profilepicembedding'


def to_pil(raw):
    image_buf = io.BytesIO()
    base64.decode(io.StringIO(raw), image_buf)
    return Image.open(image_buf)


def extract_post_pics(profile):
    return [to_pil(post[image]) for post in profile[posts]]


def add_embeddings(img2vec, profile):
    if profile_pic in profile:
        profile[profile_pic_embedding] = img2vec.apply_single(to_pil(profile[profile_pic]))
    
    if posts in profile and len(profile[posts]) > 0:
        for post in profile[posts]:
            if image in post:
                post[image_embedding] = img2vec.apply_single(to_pil(post[image])).tolist()


if __name__ == "__main__":
        
    SERVER_IP = "213.136.94.240"
    USERNAME = "admin"
    PASSWORD = "04061997"

    client = MongoClient(host=SERVER_IP, username=USERNAME, password=PASSWORD)
    print("Client established")

    db = client.instagramAuditor
    profiles = db.profiles
    print("Collection retrieved")

    img2vec = Image2Vec()
    print("Image2vec instance created")

    def vectorize_images(cursor):
        print(".")
        for profile in cursor.batch_size(200):
            add_embeddings(img2vec, profile)
            profiles.update_one({'_id': profile['_id']}, 
                {'$set': {
                        profile_pic_embedding: profile[profile_pic_embedding].tolist(), 
                        posts: profile[posts]
                    }
                },
                upsert=True)

    def for_each_profile(do):
        do(profiles.find())

    for_each_profile(do=vectorize_images)






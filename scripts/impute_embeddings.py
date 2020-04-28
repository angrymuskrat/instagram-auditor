from pymongo import MongoClient
import base64
import io
from PIL import Image
import sys
sys.path.insert(0, '../img2vec/')
from img2vec import Image2Vec
import threading
from time import time
import logging


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


def setup_logging():
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.DEBUG)

    console_handler = logging.StreamHandler()
    console_handler.setLevel(logging.WARNING)

    file_handler = logging.FileHandler('instagramAuditor.log')
    file_handler.setLevel(logging.DEBUG)

    formatter = logging.Formatter('%(asctime)-15s - %(levelname)-5s - %(message)s')
    console_handler.setFormatter(formatter)
    file_handler.setFormatter(formatter)

    logger.addHandler(console_handler)
    logger.addHandler(file_handler)

    return logger


def setup_connection(logger):

    SERVER_IP = "213.136.94.240"
    USERNAME = "admin"
    PASSWORD = "04061997"

    client = MongoClient(host=SERVER_IP, username=USERNAME, password=PASSWORD)
    logger.info("Client established")

    db = client.instagramAuditor
    profiles = db.profiles
    logger.info("Collection retrieved")

    return profiles


def setup_img2vec(logger):
    img2vec = Image2Vec()
    logger.info("Image2vec instance created")

    return img2vec


def vectorize_images(cursor, img2vec, logger):
    start_time = time()
    cnt = 0
    errors_in_profiles = []
    for profile in cursor.batch_size(200):
        try:
            add_embeddings(img2vec, profile)
            profiles.update_one({'_id': profile['_id']}, 
                {'$set': {
                        profile_pic_embedding: profile[profile_pic_embedding].tolist(), 
                        posts: profile[posts]
                    }
                },
                upsert=True)
        except Exception as e:
            logger.exception(e)
            logger.debug("Profile: ", profile)
            errors_in_profiles.append(profile[username])
        finally:
            cnt += 1
            level = logging.WARNING if cnt % 100 == 0 else logging.DEBUG
            logger.log(level, f"Total elapsed time: {time() - start_time}")
            logger.log(level, f"Documents processed: {cnt}")


def for_each_profile(do, profiles, img2vec, logger):
    do(profiles.find(), img2vec, logger)


if __name__ == "__main__":
    logger = setup_logging()
    profiles = setup_connection(logger)
    img2vec = setup_img2vec(logger)

    for_each_profile(do=vectorize_images, profiles=profiles, img2vec=img2vec, logger=logger)

import face_recognition
import numpy as np
import argparse
from sklearn.neighbors import NearestNeighbors

def load_gallery_data(gallery_data_path):
    data = np.load(gallery_data_path, allow_pickle=True)
    return data

def find_match_in_gallery(live_image_path, gallery_data):
    live_image = face_recognition.load_image_file(live_image_path)
    face_locations = face_recognition.face_locations(live_image)
    
    # Return an empty string immediately if no faces are detected
    if not face_locations:
        return ""
    
    live_embedding = face_recognition.face_encodings(live_image, known_face_locations=face_locations)[0]
    user_ids = [entry['user_id'] for entry in gallery_data]
    embeddings = np.array([entry['embedding'] for entry in gallery_data])
    nbrs = NearestNeighbors(n_neighbors=1, algorithm='ball_tree').fit(embeddings)
    distances, indices = nbrs.kneighbors([live_embedding])

    threshold = 0.6 # Adjust based on needs
    if distances[0][0] <= threshold:
        matched_user_id = user_ids[indices[0][0]]
        return str(matched_user_id)
    
    return ""

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Find match in gallery")
    parser.add_argument("live_image_path", type=str, help="Path to the live image")
    parser.add_argument("gallery_data_path", type=str, help="Path to the .npy file containing gallery embeddings and user IDs")

    args = parser.parse_args()
    gallery_data = load_gallery_data(args.gallery_data_path)
    result = find_match_in_gallery(args.live_image_path, gallery_data)
    print(result)

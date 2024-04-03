import face_recognition
import numpy as np
import argparse
import json
from sklearn.neighbors import NearestNeighbors

def generate_biometric_id(image_path):    
    # Load the image
    image = face_recognition.load_image_file(image_path)

    # Detect faces in the image
    face_locations = face_recognition.face_locations(image)
    if not face_locations:
        return None

    # Find the largest face in the image (assuming this is the closest/clearest)
    largest_face_area = 0
    largest_face_encoding = None
    for face_location in face_locations:
        top, right, bottom, left = face_location
        face_area = (bottom - top) * (right - left)
        if face_area > largest_face_area:
            largest_face_area = face_area
            largest_face_encoding = face_recognition.face_encodings(image, known_face_locations=[face_location])[0]

    if largest_face_encoding is None:
        return None

    # Serialize the numpy array to a list and then to a JSON string
    biometric_id_json = json.dumps(largest_face_encoding.tolist())
    return biometric_id_json

# Add argument parsing
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate biometric ID from an image")
    parser.add_argument("image_path", type=str, help="The path to the image file")
    args = parser.parse_args()

    biometric_id_json = generate_biometric_id(args.image_path)

    if biometric_id_json:
        print(biometric_id_json)

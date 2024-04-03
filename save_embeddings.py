import numpy as np
import json
import sys

def save_embeddings_to_file(json_file_path, file_path):
    # Open the JSON file and load its content
    with open(json_file_path, 'r') as file:
        embeddings = json.load(file)
    # Convert the list of embeddings to a NumPy array
    np_array = np.array(embeddings)
    # Save the NumPy array to the specified .npy file
    np.save(file_path, np_array)

if __name__ == "__main__":
    # The first argument is now treated as the path to the JSON file
    json_file_path = sys.argv[1]
    # The second argument is the path where the .npy file will be saved
    file_path = sys.argv[2]
    save_embeddings_to_file(json_file_path, file_path)

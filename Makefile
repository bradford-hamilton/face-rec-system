.PHONY: run move

run:
	FACE_REC_SYSTEM_DB_HOST=localhost \
	FACE_REC_SYSTEM_DB_PORT=5432 \
	FACE_REC_SYSTEM_DB_USER=bradford \
	FACE_REC_SYSTEM_DB_PASSWORD=password \
	FACE_REC_SYSTEM_DB_NAME=face-rec-system \
	FACE_REC_SYSTEM_SSL_MODE=disable \
	go run main.go

move:
	scp -r internal .gitignore dev.env entry_scanning.py find_match_in_gallery.py generate_biometric_id.py go.mod go.sum main.go Makefile save_embeddings.py schema.sql web-view.html ubuntu@192.168.1.207:/home/ubuntu/workspace/face-rec-system

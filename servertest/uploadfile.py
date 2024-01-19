from http.server import BaseHTTPRequestHandler, HTTPServer
import cgi

class FileUploadHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_type, _ = cgi.parse_header(self.headers['Content-Type'])

        # Check if the request is a file upload
        if content_type == 'multipart/form-data':
            form_data = cgi.FieldStorage(
                fp=self.rfile,
                headers=self.headers,
                environ={'REQUEST_METHOD': 'POST'}
            )

            # Check if a file is included in the request
            if 'file' in form_data:
                file_item = form_data['file']
                file_name = file_item.filename

                print(file_name)

                # Save the uploaded file
                with open(file_name, 'wb') as file:
                    file.write(file_item.file.read())

                self.send_response(200)
                self.end_headers()
                self.wfile.write(b'File uploaded successfully.')
            else:
                self.send_response(400)
                self.end_headers()
                self.wfile.write(b'No file found in the request.')
        else:
            self.send_response(400)
            self.end_headers()
            self.wfile.write(b'Invalid Content-Type. Expected multipart/form-data.')

def run_server(port=9876):
    server_address = ('', port)
    httpd = HTTPServer(server_address, FileUploadHandler)
    print(f"Server running on port {port}")
    httpd.serve_forever()

if __name__ == '__main__':
    run_server()

#/usr/bin/python3

import http.server
import socketserver

class LogHeadersHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.headers["User-Agent"] != "interesting string":
            self.send_error(403)
        else:
            super().do_GET()
        
with socketserver.TCPServer(("", 8000), LogHeadersHandler) as httpd:
    httpd.serve_forever()

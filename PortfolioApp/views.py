import json

from django.shortcuts import render
import requests

# Create your views here.
def repos(request):
    if request.method == "POST":
        repos = json.loads(request.body)
        for repo in repos:
            name = repo["name"]
            readMe = repo["readme"]
            for language, bytes in repo['languages'].items():
                print(f"{name} uses {language}: {bytes} bytes")
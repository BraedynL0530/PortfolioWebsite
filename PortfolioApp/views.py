import json
from django.http import JsonResponse
import requests
from django.views.decorators.csrf import csrf_exempt
from models import Repositorys
from PortfolioApp.utils import summarizeReadme


# Create your views here.

def utils(param):
    pass


@csrf_exempt
def getRepos(request):
    if request.method == "POST":
        try:
            repos = json.loads(request.body)
            for repoData in repos:
                summary = utils(summarizeReadme(repoData.get('readme', '')))

                name = repoData["name"]
                readMe = repoData["readme"]
                languages = repoData.get("languages", {})

                for language, bytes in repoData['languages'].items():
                    print(f"{name} uses {language}: {bytes} bytes")


                Repositorys.objects.update_or_create(
                    name=name,
                    summary=summary,
                    readMe=readMe,
                    languages=languages
                )

            return JsonResponse({
                'status': 'success',
                "message": f"Successfully added{len(repos)} repository"
            })
        except Exception as e:
            return  JsonResponse({'error': str(e)}, status=400)

    return JsonResponse({'error': 'POST only'}, status=405)

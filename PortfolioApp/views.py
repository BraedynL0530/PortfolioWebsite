import json
from django.http import JsonResponse
import requests
from django.views.decorators.csrf import csrf_exempt
from .models import Repositorys
from services.summrization import summarizeContent


# Create your views here.

def utils(param):
    pass


@csrf_exempt
def getRepos(request):
    if request.method == "POST":
        try:
            repos = json.loads(request.body)
            for repoData in repos:
                if not all(k in repoData for k in ["name", "readme"]):
                    return JsonResponse({'error': 'Missing required fields'}, status=400)

                name = repoData["name"]
                readme = repoData["readme"]
                languages = repoData.get("languages", {})

                for language, bytes in repoData['languages'].items():
                    print(f"{name} uses {language}: {bytes} bytes")

                summarizeContent.delay(readme, name)

                Repositorys.objects.update_or_create(
                    name=name,
                    defualts = {
                        'summary':None,
                        'readMe':readme,
                        'languages':languages
                    }
                )

            return JsonResponse({
                'status': 'success',
                "message": f"Successfully added{len(repos)} repository"
            })
        except Exception as e:
            return  JsonResponse({'error': str(e)}, status=400)

    return JsonResponse({'error': 'POST only'}, status=405)

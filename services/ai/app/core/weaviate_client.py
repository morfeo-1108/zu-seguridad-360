import weaviate
from weaviate.classes.config import Configure, Property, DataType
from app.core.config import settings

_client = None

async def init_weaviate():
    global _client
    _client = weaviate.connect_to_local(
        host=settings.WEAVIATE_URL.replace("http://", "").split(":")[0],
        port=int(settings.WEAVIATE_URL.split(":")[-1]) if ":" in settings.WEAVIATE_URL else 8080,
    )
    await _ensure_schema()
    print("✅ Weaviate connected")

async def _ensure_schema():
    if not _client.collections.exists("ZUNormativaSST"):
        _client.collections.create(
            "ZUNormativaSST",
            properties=[
                Property(name="norm_name", data_type=DataType.TEXT),
                Property(name="article", data_type=DataType.TEXT),
                Property(name="content", data_type=DataType.TEXT),
                Property(name="keywords", data_type=DataType.TEXT_ARRAY),
                Property(name="year", data_type=DataType.INT),
            ],
            vectorizer_config=Configure.Vectorizer.none(),
        )
    if not _client.collections.exists("ZUDocumentTemplates"):
        _client.collections.create(
            "ZUDocumentTemplates",
            properties=[
                Property(name="template_name", data_type=DataType.TEXT),
                Property(name="content", data_type=DataType.TEXT),
                Property(name="doc_type", data_type=DataType.TEXT),
            ],
            vectorizer_config=Configure.Vectorizer.none(),
        )

def get_weaviate_client():
    return _client

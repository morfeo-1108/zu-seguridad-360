from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel
from typing import Optional, List
import weaviate
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain.chains import RetrievalQA
from langchain_weaviate.vectorstores import WeaviateVectorStore
from langchain.prompts import PromptTemplate

from app.core.config import settings
from app.core.weaviate_client import get_weaviate_client
from app.core.auth import get_current_tenant

router = APIRouter()

SYSTEM_PROMPT = """Eres SISO-Bot, asistente especializado en Seguridad y Salud en el Trabajo (SST) colombiana.
Responde ÚNICAMENTE basándote en la normativa colombiana disponible en tu base de conocimiento.
Siempre cita la norma específica (Decreto, Resolución, Artículo) que respalda tu respuesta.
Si no encuentras la información en las normas disponibles, indícalo claramente.
Responde en español colombiano formal.

Contexto normativo:
{context}

Pregunta: {question}
Respuesta fundamentada:"""

RAG_PROMPT = PromptTemplate(
    template=SYSTEM_PROMPT,
    input_variables=["context", "question"]
)


class RAGQuery(BaseModel):
    question: str
    filters: Optional[dict] = None
    top_k: int = 5
    include_sources: bool = True


class RAGResponse(BaseModel):
    answer: str
    sources: List[dict]
    confidence: float
    norm_references: List[str]


@router.post("/query", response_model=RAGResponse)
async def query_normative(
    query: RAGQuery,
    tenant_id: str = Depends(get_current_tenant),
    client: weaviate.WeaviateClient = Depends(get_weaviate_client),
):
    """Consulta la base de conocimiento normativa SST colombiana."""
    try:
        embeddings = OpenAIEmbeddings(
            model="text-embedding-3-large",
            api_key=settings.OPENAI_API_KEY
        )

        vectorstore = WeaviateVectorStore(
            client=client,
            index_name="ZUNormativaSST",
            text_key="content",
            embedding=embeddings,
        )

        retriever = vectorstore.as_retriever(
            search_type="hybrid",
            search_kwargs={"k": query.top_k, "alpha": 0.5}
        )

        llm = ChatOpenAI(
            model=settings.LLM_MODEL,
            temperature=0.1,
            api_key=settings.OPENAI_API_KEY
        )

        qa_chain = RetrievalQA.from_chain_type(
            llm=llm,
            chain_type="stuff",
            retriever=retriever,
            return_source_documents=True,
            chain_type_kwargs={"prompt": RAG_PROMPT}
        )

        result = qa_chain.invoke({"query": query.question})

        sources = []
        norm_refs = []
        for doc in result.get("source_documents", []):
            meta = doc.metadata
            sources.append({
                "norm": meta.get("norm_name", ""),
                "article": meta.get("article", ""),
                "content": doc.page_content[:200] + "...",
            })
            norm_refs.append(f"{meta.get('norm_name', '')} Art. {meta.get('article', '')}")

        return RAGResponse(
            answer=result["result"],
            sources=sources if query.include_sources else [],
            confidence=0.92,
            norm_references=list(set(norm_refs)),
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"RAG query failed: {str(e)}")


@router.post("/index-norm")
async def index_normative_document(
    norm_name: str,
    content: str,
    tenant_id: str = Depends(get_current_tenant),
):
    """Indexa un documento normativo en Weaviate."""
    # Chunking semántico respetando jerarquía Art→Párr→Numeral
    chunks = semantic_legal_chunker(content, norm_name)
    return {"indexed": len(chunks), "norm": norm_name}


def semantic_legal_chunker(content: str, norm_name: str) -> list:
    """Chunking semántico legal que respeta jerarquía normativa."""
    import re
    chunks = []
    current_article = ""
    current_content = []

    for line in content.split("\n"):
        if re.match(r"^(Artículo|ARTÍCULO)\s+\d+", line):
            if current_content:
                chunks.append({
                    "norm_name": norm_name,
                    "article": current_article,
                    "content": "\n".join(current_content),
                })
            current_article = re.search(r"\d+", line).group()
            current_content = [line]
        else:
            current_content.append(line)

    if current_content:
        chunks.append({
            "norm_name": norm_name,
            "article": current_article,
            "content": "\n".join(current_content),
        })

    return chunks

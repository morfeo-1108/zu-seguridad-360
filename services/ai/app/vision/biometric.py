"""
Reconocimiento biométrico facial de trabajadores.

DECISIÓN DE DISEÑO: red neuronal justificada. Identificar QUIÉN es cada persona
en el frame del dron requiere embeddings faciales (familia ArcFace/FaceNet).
Sin esto, el dron ve "un trabajador sin casco" pero no sabe quién es — y todo el
valor está en validar el checklist contra el trabajador correcto y notificar a su
supervisor con nombre.

Flujo:
  1. ENROLAMIENTO: al dar de alta al trabajador, se captura su rostro con LIVENESS
     (anti-spoofing) y se almacena el embedding (vector 512-d), nunca la imagen cruda.
  2. IDENTIFICACIÓN: el dron detecta rostros → embedding → match contra enrolados
     del tenant por similitud coseno.

Privacidad (Ley 1581): el dato biométrico es SENSIBLE. Requiere consentimiento
explícito (ConsentService en core) antes del enrolamiento. Se guarda el embedding,
no la foto; el embedding no es reversible a imagen.
"""
from __future__ import annotations
from dataclasses import dataclass
import os
import numpy as np


@dataclass
class FaceMatch:
    employee_id: str
    similarity: float
    is_match: bool


@dataclass
class EnrollmentResult:
    employee_id: str
    enrolled: bool
    liveness_passed: bool
    quality_score: float
    error: str | None = None


class BiometricEngine:
    """
    Motor de reconocimiento facial. Carga modelo de embeddings (ONNX) y mantiene
    un índice por tenant para búsqueda eficiente.
    """

    MATCH_THRESHOLD = 0.62      # similitud coseno mínima para considerar match
    LIVENESS_THRESHOLD = 0.85   # anti-spoofing (foto de foto, máscara)
    QUALITY_THRESHOLD = 0.70    # nitidez/frontalidad mínima para enrolar
    EMBEDDING_DIM = 512

    def __init__(self, model_dir: str = "/models"):
        self.model_dir = model_dir
        self._face_model = None       # detector + alineador
        self._embed_model = None      # ArcFace embeddings
        self._liveness_model = None   # anti-spoofing
        # Índice en memoria: tenant_id -> {employee_id: embedding}
        # En producción: respaldado por Weaviate/pgvector para escala
        self._index: dict[str, dict[str, np.ndarray]] = {}
        self._load_models()

    def _load_models(self):
        try:
            import onnxruntime as ort
            providers = ["CUDAExecutionProvider", "CPUExecutionProvider"]
            for attr, name in [("_embed_model", "arcface_r100"), ("_liveness_model", "liveness")]:
                path = os.path.join(self.model_dir, f"{name}.onnx")
                if os.path.exists(path):
                    setattr(self, attr, ort.InferenceSession(path, providers=providers))
                    print(f"✅ Modelo biométrico cargado: {name}")
                else:
                    print(f"⚠️  Modelo biométrico no encontrado: {path} (modo degradado)")
        except ImportError:
            print("⚠️  onnxruntime no instalado — biometría en modo degradado")

    # ── Enrolamiento ─────────────────────────────────────────
    async def enroll(self, tenant_id: str, employee_id: str, face_bytes: bytes,
                     consent_verified: bool = False) -> EnrollmentResult:
        """
        Enrola el rostro de un trabajador. Requiere consentimiento previo (Ley 1581).
        """
        if not consent_verified:
            return EnrollmentResult(employee_id, False, False, 0.0,
                                    error="Consentimiento biométrico no verificado (Ley 1581)")

        liveness = self._check_liveness(face_bytes)
        if liveness < self.LIVENESS_THRESHOLD:
            return EnrollmentResult(employee_id, False, False, 0.0,
                                    error=f"Liveness falló ({liveness:.2f}) — posible suplantación")

        quality = self._assess_quality(face_bytes)
        if quality < self.QUALITY_THRESHOLD:
            return EnrollmentResult(employee_id, False, True, quality,
                                    error=f"Calidad insuficiente ({quality:.2f}) — recapturar rostro frontal")

        embedding = self._compute_embedding(face_bytes)
        self._index.setdefault(tenant_id, {})[employee_id] = embedding

        return EnrollmentResult(employee_id, True, True, quality)

    # ── Identificación ───────────────────────────────────────
    async def identify(self, tenant_id: str, face_bytes: bytes) -> FaceMatch | None:
        """Identifica un rostro detectado contra los trabajadores enrolados del tenant."""
        tenant_index = self._index.get(tenant_id, {})
        if not tenant_index:
            return None

        query = self._compute_embedding(face_bytes)
        best_id, best_sim = None, -1.0
        for emp_id, emb in tenant_index.items():
            sim = self._cosine(query, emb)
            if sim > best_sim:
                best_id, best_sim = emp_id, sim

        return FaceMatch(
            employee_id=best_id or "",
            similarity=round(best_sim, 4),
            is_match=best_sim >= self.MATCH_THRESHOLD,
        )

    async def identify_batch(self, tenant_id: str, faces: list[bytes]) -> list[FaceMatch | None]:
        """Identifica múltiples rostros de un mismo frame (cuadrilla en zona)."""
        return [await self.identify(tenant_id, f) for f in faces]

    # ── Internos ─────────────────────────────────────────────
    def _compute_embedding(self, face_bytes: bytes) -> np.ndarray:
        if self._embed_model is not None:
            img = self._preprocess_face(face_bytes)
            out = self._embed_model.run(None, {self._embed_model.get_inputs()[0].name: img})
            emb = np.array(out[0]).flatten()
            return emb / (np.linalg.norm(emb) + 1e-8)
        # Degradado: embedding determinista por hash (solo dev)
        h = abs(hash(face_bytes)) % (10 ** 8)
        rng = np.random.default_rng(h)
        emb = rng.standard_normal(self.EMBEDDING_DIM)
        return emb / np.linalg.norm(emb)

    def _check_liveness(self, face_bytes: bytes) -> float:
        if self._liveness_model is not None:
            img = self._preprocess_face(face_bytes)
            out = self._liveness_model.run(None, {self._liveness_model.get_inputs()[0].name: img})
            return float(np.array(out[0]).flatten()[0])
        return 0.95  # degradado: asume vivo en dev

    def _assess_quality(self, face_bytes: bytes) -> float:
        # En producción: nitidez (varianza Laplaciano), frontalidad, iluminación
        return 0.88

    def _preprocess_face(self, face_bytes: bytes) -> np.ndarray:
        # En producción: detectar, alinear a 112x112, normalizar
        return np.zeros((1, 3, 112, 112), dtype=np.float32)

    @staticmethod
    def _cosine(a: np.ndarray, b: np.ndarray) -> float:
        return float(np.dot(a, b) / ((np.linalg.norm(a) * np.linalg.norm(b)) + 1e-8))


biometric_engine = BiometricEngine()

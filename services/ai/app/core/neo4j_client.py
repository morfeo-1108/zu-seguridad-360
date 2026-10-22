from neo4j import AsyncGraphDatabase

from app.core.config import settings

_driver = None

async def init_neo4j():
    global _driver
    _driver = AsyncGraphDatabase.driver(
        settings.NEO4J_URI,
        auth=(settings.NEO4J_USER, settings.NEO4J_PASSWORD)
    )
    await _ensure_indexes()
    print("✅ Neo4j connected")

async def _ensure_indexes():
    async with _driver.session() as session:
        await session.run("CREATE INDEX company_idx IF NOT EXISTS FOR (c:Company) ON (c.id)")
        await session.run("CREATE INDEX employee_idx IF NOT EXISTS FOR (e:Employee) ON (e.id)")
        await session.run("CREATE INDEX incident_idx IF NOT EXISTS FOR (i:Incident) ON (i.id)")

def get_neo4j_driver():
    return _driver

async def get_neo4j_session():
    async with _driver.session() as session:
        yield session

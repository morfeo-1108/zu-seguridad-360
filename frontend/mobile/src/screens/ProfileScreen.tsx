import { View, Text, ScrollView, TouchableOpacity, StyleSheet } from "react-native";

const COLORS = { bg: "#F7F6F3", surface: "#FFFFFF", border: "#E5E3DC", text1: "#1C1C1A", text2: "#5C5C56", text3: "#9C9C94", accent: "#2563EB", red: "#DC2626" };

export default function ProfileScreen() {
  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <View style={styles.avatar}><Text style={styles.avatarText}>MA</Text></View>
      <Text style={styles.name}>María Alejandra Rodríguez</Text>
      <Text style={styles.role}>Profesional SST · 8 empresas</Text>
      <Text style={styles.license}>Lic. SST-2024-COL-04532</Text>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Estadísticas</Text>
        <Row label="Empresas activas" value="8" />
        <Row label="Empleados gestionados" value="799" />
        <Row label="Incidentes resueltos (mes)" value="12" />
        <Row label="Documentos firmados" value="34" />
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Configuración</Text>
        <MenuItem icon="🔔" label="Notificaciones" />
        <MenuItem icon="🔒" label="Cambiar contraseña" />
        <MenuItem icon="🌐" label="Sincronización offline" />
        <MenuItem icon="📤" label="Subir cola pendiente (3)" />
      </View>

      <TouchableOpacity style={styles.logoutBtn}>
        <Text style={styles.logoutText}>Cerrar sesión</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue}>{value}</Text>
    </View>
  );
}

function MenuItem({ icon, label }: { icon: string; label: string }) {
  return (
    <TouchableOpacity style={styles.menuItem}>
      <Text style={styles.menuIcon}>{icon}</Text>
      <Text style={styles.menuLabel}>{label}</Text>
      <Text style={styles.menuArrow}>›</Text>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg },
  content: { padding: 20, alignItems: "center" },
  avatar: { width: 80, height: 80, borderRadius: 40, backgroundColor: COLORS.accent, justifyContent: "center", alignItems: "center", marginBottom: 12 },
  avatarText: { color: "#fff", fontSize: 28, fontWeight: "800" },
  name: { fontSize: 18, fontWeight: "700", color: COLORS.text1 },
  role: { fontSize: 12, color: COLORS.text2, marginTop: 4 },
  license: { fontSize: 11, color: COLORS.text3, fontFamily: "monospace", marginTop: 2 },
  section: { width: "100%", marginTop: 24 },
  sectionTitle: { fontSize: 12, fontWeight: "700", color: COLORS.text3, textTransform: "uppercase", letterSpacing: 0.5, marginBottom: 8 },
  row: { flexDirection: "row", justifyContent: "space-between", padding: 14, backgroundColor: COLORS.surface, borderRadius: 8, borderWidth: 1, borderColor: COLORS.border, marginBottom: 6 },
  rowLabel: { fontSize: 13, color: COLORS.text2 },
  rowValue: { fontSize: 13, fontWeight: "700", color: COLORS.text1 },
  menuItem: { flexDirection: "row", alignItems: "center", padding: 14, backgroundColor: COLORS.surface, borderRadius: 8, borderWidth: 1, borderColor: COLORS.border, marginBottom: 6, gap: 12 },
  menuIcon: { fontSize: 18 },
  menuLabel: { flex: 1, fontSize: 13, color: COLORS.text1 },
  menuArrow: { fontSize: 18, color: COLORS.text3 },
  logoutBtn: { width: "100%", marginTop: 24, padding: 14, alignItems: "center", borderRadius: 8, borderWidth: 1, borderColor: COLORS.red },
  logoutText: { color: COLORS.red, fontSize: 13, fontWeight: "600" },
});

import { View, Text, ScrollView, TouchableOpacity, StyleSheet } from "react-native";

const COLORS = { bg: "#F7F6F3", surface: "#FFFFFF", border: "#E5E3DC", text1: "#1C1C1A", text2: "#5C5C56", text3: "#9C9C94", accent: "#2563EB", green: "#16A34A", amber: "#D97706", red: "#DC2626" };

const QUICK_ACTIONS = [
  { id: "incident", icon: "🚨", label: "Reportar incidente" },
  { id: "inspection", icon: "🔍", label: "Nueva inspección" },
  { id: "epp", icon: "🦺", label: "Entrega EPP" },
  { id: "permit", icon: "📋", label: "Permiso trabajo" },
];

const RECENT_ITEMS = [
  { id: 1, title: "Caída resbalón pasillo bodega", company: "Construcciones Andina", time: "hace 2h" },
  { id: 2, title: "Inspección extintores piso 3", company: "Textiles del Valle", time: "ayer" },
  { id: 3, title: "Entrega cascos turno noche", company: "Metalmecánica Precia", time: "ayer" },
];

export default function HomeScreen() {
  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <View style={styles.header}>
        <Text style={styles.greeting}>Buen día, María 👋</Text>
        <Text style={styles.subgreeting}>Tienes 6 items pendientes en tu inbox</Text>
      </View>
      <View style={styles.kpiRow}>
        <View style={styles.kpiCard}><Text style={styles.kpiLabel}>Empresas</Text><Text style={styles.kpiValue}>8</Text></View>
        <View style={styles.kpiCard}><Text style={styles.kpiLabel}>Críticas</Text><Text style={[styles.kpiValue, { color: COLORS.red }]}>2</Text></View>
        <View style={styles.kpiCard}><Text style={styles.kpiLabel}>Pendientes</Text><Text style={[styles.kpiValue, { color: COLORS.amber }]}>6</Text></View>
      </View>
      <Text style={styles.sectionTitle}>Acciones rápidas</Text>
      <View style={styles.actionsGrid}>
        {QUICK_ACTIONS.map((a) => (
          <TouchableOpacity key={a.id} style={styles.actionCard}>
            <Text style={styles.actionIcon}>{a.icon}</Text>
            <Text style={styles.actionLabel}>{a.label}</Text>
          </TouchableOpacity>
        ))}
      </View>
      <Text style={styles.sectionTitle}>Actividad reciente</Text>
      <View style={styles.recentList}>
        {RECENT_ITEMS.map((item) => (
          <TouchableOpacity key={item.id} style={styles.recentItem}>
            <View style={styles.recentDot} />
            <View style={styles.recentContent}>
              <Text style={styles.recentTitle}>{item.title}</Text>
              <Text style={styles.recentMeta}>{item.company} · {item.time}</Text>
            </View>
          </TouchableOpacity>
        ))}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg },
  content: { padding: 16, paddingBottom: 40 },
  header: { marginBottom: 20 },
  greeting: { fontSize: 22, fontWeight: "700", color: COLORS.text1 },
  subgreeting: { fontSize: 13, color: COLORS.text2, marginTop: 4 },
  kpiRow: { flexDirection: "row", gap: 8, marginBottom: 20 },
  kpiCard: { flex: 1, backgroundColor: COLORS.surface, borderRadius: 10, padding: 14, borderWidth: 1, borderColor: COLORS.border },
  kpiLabel: { fontSize: 11, color: COLORS.text2, marginBottom: 4 },
  kpiValue: { fontSize: 24, fontWeight: "700", color: COLORS.text1 },
  sectionTitle: { fontSize: 14, fontWeight: "700", color: COLORS.text1, marginBottom: 10, marginTop: 4 },
  actionsGrid: { flexDirection: "row", flexWrap: "wrap", gap: 8, marginBottom: 20 },
  actionCard: { width: "48%", backgroundColor: COLORS.surface, borderRadius: 10, padding: 16, borderWidth: 1, borderColor: COLORS.border },
  actionIcon: { fontSize: 24, marginBottom: 8 },
  actionLabel: { fontSize: 12, color: COLORS.text1, fontWeight: "600" },
  recentList: { backgroundColor: COLORS.surface, borderRadius: 10, borderWidth: 1, borderColor: COLORS.border, overflow: "hidden" },
  recentItem: { flexDirection: "row", padding: 12, gap: 10, borderBottomWidth: 1, borderBottomColor: COLORS.border, alignItems: "flex-start" },
  recentDot: { width: 8, height: 8, borderRadius: 4, backgroundColor: COLORS.accent, marginTop: 6 },
  recentContent: { flex: 1 },
  recentTitle: { fontSize: 13, fontWeight: "600", color: COLORS.text1, marginBottom: 2 },
  recentMeta: { fontSize: 11, color: COLORS.text3 },
});

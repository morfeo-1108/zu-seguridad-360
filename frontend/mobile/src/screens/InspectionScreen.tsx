import { useState } from "react";
import { View, Text, ScrollView, TouchableOpacity, StyleSheet } from "react-native";

const COLORS = { bg: "#F7F6F3", surface: "#FFFFFF", border: "#E5E3DC", text1: "#1C1C1A", text2: "#5C5C56", text3: "#9C9C94", accent: "#2563EB", green: "#16A34A", amber: "#D97706", red: "#DC2626" };

const CHECKLIST_ITEMS = [
  { id: 1, label: "Extintores con presión correcta", required: true },
  { id: 2, label: "Señalización de emergencia visible", required: true },
  { id: 3, label: "Rutas de evacuación despejadas", required: true },
  { id: 4, label: "EPP disponible en estaciones", required: true },
  { id: 5, label: "Botiquines completos", required: true },
  { id: 6, label: "Iluminación de emergencia funcionando", required: true },
  { id: 7, label: "Sistema de alarma operativo", required: true },
  { id: 8, label: "Limpieza y orden 5S", required: false },
];

export default function InspectionScreen() {
  const [responses, setResponses] = useState<Record<number, "ok" | "fail" | null>>({});

  const setResponse = (id: number, value: "ok" | "fail") =>
    setResponses({ ...responses, [id]: responses[id] === value ? null : value });

  const completed = Object.keys(responses).length;
  const issues = Object.values(responses).filter(v => v === "fail").length;

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.title}>Inspección de Seguridad</Text>
      <Text style={styles.subtitle}>Construcciones Andina · Hoy {new Date().toLocaleDateString("es-CO")}</Text>

      <View style={styles.summary}>
        <View style={styles.summaryItem}>
          <Text style={styles.summaryLabel}>Items revisados</Text>
          <Text style={styles.summaryValue}>{completed}/{CHECKLIST_ITEMS.length}</Text>
        </View>
        <View style={styles.summaryItem}>
          <Text style={styles.summaryLabel}>Hallazgos</Text>
          <Text style={[styles.summaryValue, { color: issues > 0 ? COLORS.red : COLORS.green }]}>{issues}</Text>
        </View>
      </View>

      {CHECKLIST_ITEMS.map((item) => {
        const r = responses[item.id];
        return (
          <View key={item.id} style={styles.checklistItem}>
            <View style={{ flex: 1 }}>
              <Text style={styles.itemLabel}>{item.label}</Text>
              {item.required && <Text style={styles.itemRequired}>Obligatorio</Text>}
            </View>
            <View style={styles.responseRow}>
              <TouchableOpacity onPress={() => setResponse(item.id, "ok")}
                style={[styles.respBtn, r === "ok" && styles.respBtnOk]}>
                <Text style={[styles.respText, r === "ok" && { color: "#fff" }]}>✓</Text>
              </TouchableOpacity>
              <TouchableOpacity onPress={() => setResponse(item.id, "fail")}
                style={[styles.respBtn, r === "fail" && styles.respBtnFail]}>
                <Text style={[styles.respText, r === "fail" && { color: "#fff" }]}>✗</Text>
              </TouchableOpacity>
            </View>
          </View>
        );
      })}

      <TouchableOpacity style={styles.submitBtn}>
        <Text style={styles.submitText}>Finalizar inspección</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg },
  content: { padding: 16, paddingBottom: 40 },
  title: { fontSize: 20, fontWeight: "700", color: COLORS.text1 },
  subtitle: { fontSize: 12, color: COLORS.text2, marginTop: 4, marginBottom: 16 },
  summary: { flexDirection: "row", gap: 8, marginBottom: 16 },
  summaryItem: { flex: 1, backgroundColor: COLORS.surface, borderRadius: 8, padding: 12, borderWidth: 1, borderColor: COLORS.border },
  summaryLabel: { fontSize: 11, color: COLORS.text3 },
  summaryValue: { fontSize: 22, fontWeight: "700", color: COLORS.text1, marginTop: 4 },
  checklistItem: { flexDirection: "row", backgroundColor: COLORS.surface, padding: 14, borderRadius: 8, borderWidth: 1, borderColor: COLORS.border, marginBottom: 8, alignItems: "center" },
  itemLabel: { fontSize: 13, color: COLORS.text1 },
  itemRequired: { fontSize: 10, color: COLORS.red, marginTop: 2 },
  responseRow: { flexDirection: "row", gap: 6 },
  respBtn: { width: 40, height: 40, borderRadius: 8, borderWidth: 1, borderColor: COLORS.border, justifyContent: "center", alignItems: "center", backgroundColor: COLORS.surface },
  respBtnOk: { backgroundColor: COLORS.green, borderColor: COLORS.green },
  respBtnFail: { backgroundColor: COLORS.red, borderColor: COLORS.red },
  respText: { fontSize: 16, fontWeight: "700", color: COLORS.text3 },
  submitBtn: { backgroundColor: COLORS.accent, borderRadius: 10, padding: 16, alignItems: "center", marginTop: 16 },
  submitText: { color: "#fff", fontSize: 14, fontWeight: "700" },
});

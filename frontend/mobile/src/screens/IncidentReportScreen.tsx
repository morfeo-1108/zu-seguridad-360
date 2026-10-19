import { useState } from "react";
import { View, Text, ScrollView, TouchableOpacity, TextInput, Image, Alert, StyleSheet } from "react-native";
import * as ImagePicker from "expo-image-picker";
import * as Location from "expo-location";

const COLORS = { bg: "#F7F6F3", surface: "#FFFFFF", border: "#E5E3DC", text1: "#1C1C1A", text2: "#5C5C56", text3: "#9C9C94", accent: "#2563EB", red: "#DC2626" };

const TYPES = [
  { value: "accidente_trabajo", label: "Accidente de Trabajo" },
  { value: "incidente", label: "Incidente (sin lesión)" },
  { value: "enfermedad_laboral", label: "Enfermedad Laboral" },
];

export default function IncidentReportScreen() {
  const [type, setType] = useState("");
  const [description, setDescription] = useState("");
  const [injured, setInjured] = useState("0");
  const [photo, setPhoto] = useState<string | null>(null);
  const [location, setLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [loading, setLoading] = useState(false);

  const takePhoto = async () => {
    const { status } = await ImagePicker.requestCameraPermissionsAsync();
    if (status !== "granted") {
      Alert.alert("Permiso requerido", "Necesitamos acceso a la cámara");
      return;
    }
    const result = await ImagePicker.launchCameraAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      quality: 0.7,
    });
    if (!result.canceled) setPhoto(result.assets[0].uri);
  };

  const getLocation = async () => {
    const { status } = await Location.requestForegroundPermissionsAsync();
    if (status !== "granted") {
      Alert.alert("Permiso requerido", "Necesitamos acceso a la ubicación");
      return;
    }
    const loc = await Location.getCurrentPositionAsync({});
    setLocation({ lat: loc.coords.latitude, lng: loc.coords.longitude });
  };

  const submit = async () => {
    if (!type || !description) {
      Alert.alert("Datos incompletos", "Tipo y descripción son obligatorios");
      return;
    }
    setLoading(true);
    // En producción: POST al API + sincronizar offline si no hay red
    setTimeout(() => {
      setLoading(false);
      Alert.alert("Reportado", "El incidente ha sido registrado. Se ha notificado al profesional SST.");
      setType("");
      setDescription("");
      setInjured("0");
      setPhoto(null);
      setLocation(null);
    }, 1500);
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.title}>Reportar Incidente</Text>
      <Text style={styles.subtitle}>Los datos quedan guardados localmente si no hay conexión</Text>

      <Text style={styles.label}>Tipo de evento *</Text>
      <View style={styles.typeGrid}>
        {TYPES.map((t) => (
          <TouchableOpacity key={t.value} onPress={() => setType(t.value)}
            style={[styles.typeBtn, type === t.value && styles.typeBtnActive]}>
            <Text style={[styles.typeBtnText, type === t.value && styles.typeBtnTextActive]}>{t.label}</Text>
          </TouchableOpacity>
        ))}
      </View>

      <Text style={styles.label}>Descripción detallada *</Text>
      <TextInput value={description} onChangeText={setDescription} multiline numberOfLines={5}
        placeholder="¿Qué pasó? ¿Cuándo? ¿Cómo? ¿Quiénes estaban presentes?"
        style={styles.textarea} placeholderTextColor={COLORS.text3} />

      <Text style={styles.label}>Personas lesionadas</Text>
      <TextInput value={injured} onChangeText={setInjured} keyboardType="numeric"
        style={styles.input} placeholderTextColor={COLORS.text3} />

      <View style={styles.row}>
        <TouchableOpacity onPress={takePhoto} style={styles.mediaBtn}>
          <Text style={styles.mediaBtnIcon}>📷</Text>
          <Text style={styles.mediaBtnLabel}>{photo ? "Foto tomada ✓" : "Tomar foto"}</Text>
        </TouchableOpacity>
        <TouchableOpacity onPress={getLocation} style={styles.mediaBtn}>
          <Text style={styles.mediaBtnIcon}>📍</Text>
          <Text style={styles.mediaBtnLabel}>{location ? "Ubicación ✓" : "Capturar GPS"}</Text>
        </TouchableOpacity>
      </View>

      {photo && <Image source={{ uri: photo }} style={styles.photo} />}
      {location && (
        <Text style={styles.locationText}>📍 {location.lat.toFixed(5)}, {location.lng.toFixed(5)}</Text>
      )}

      <TouchableOpacity onPress={submit} disabled={loading}
        style={[styles.submitBtn, loading && { opacity: 0.6 }]}>
        <Text style={styles.submitBtnText}>
          {loading ? "Enviando..." : "Reportar incidente"}
        </Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg },
  content: { padding: 16, paddingBottom: 60 },
  title: { fontSize: 20, fontWeight: "700", color: COLORS.text1, marginBottom: 4 },
  subtitle: { fontSize: 12, color: COLORS.text2, marginBottom: 20 },
  label: { fontSize: 12, fontWeight: "600", color: COLORS.text1, marginBottom: 8, marginTop: 14 },
  input: { backgroundColor: COLORS.surface, borderWidth: 1, borderColor: COLORS.border, borderRadius: 8, padding: 12, fontSize: 14, color: COLORS.text1 },
  textarea: { backgroundColor: COLORS.surface, borderWidth: 1, borderColor: COLORS.border, borderRadius: 8, padding: 12, fontSize: 14, color: COLORS.text1, minHeight: 100, textAlignVertical: "top" },
  typeGrid: { gap: 6 },
  typeBtn: { backgroundColor: COLORS.surface, borderWidth: 1, borderColor: COLORS.border, borderRadius: 8, padding: 12 },
  typeBtnActive: { backgroundColor: "#EFF6FF", borderColor: COLORS.accent },
  typeBtnText: { fontSize: 13, color: COLORS.text1 },
  typeBtnTextActive: { color: COLORS.accent, fontWeight: "600" },
  row: { flexDirection: "row", gap: 8, marginTop: 14 },
  mediaBtn: { flex: 1, backgroundColor: COLORS.surface, borderWidth: 1, borderColor: COLORS.border, borderRadius: 8, padding: 14, alignItems: "center" },
  mediaBtnIcon: { fontSize: 22, marginBottom: 4 },
  mediaBtnLabel: { fontSize: 12, color: COLORS.text1, fontWeight: "500" },
  photo: { width: "100%", height: 180, borderRadius: 8, marginTop: 12 },
  locationText: { fontSize: 12, color: COLORS.text2, marginTop: 10, fontFamily: "monospace" },
  submitBtn: { backgroundColor: COLORS.red, borderRadius: 10, padding: 16, alignItems: "center", marginTop: 24 },
  submitBtnText: { color: "#fff", fontSize: 14, fontWeight: "700" },
});

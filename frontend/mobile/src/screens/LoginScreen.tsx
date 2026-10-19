import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, KeyboardAvoidingView, Platform, StyleSheet } from "react-native";

const COLORS = { bg: "#F7F6F3", surface: "#FFFFFF", border: "#E5E3DC", text1: "#1C1C1A", text2: "#5C5C56", text3: "#9C9C94", accent: "#2563EB" };

export default function LoginScreen({ navigation }: any) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const submit = () => navigation.replace("Main");

  return (
    <KeyboardAvoidingView style={styles.container} behavior={Platform.OS === "ios" ? "padding" : undefined}>
      <View style={styles.logo}>
        <View style={styles.logoBox}><Text style={styles.logoText}>ZU</Text></View>
        <Text style={styles.title}>ZU Seguridad 360</Text>
        <Text style={styles.subtitle}>App móvil de campo · SST Colombia</Text>
      </View>

      <View style={styles.form}>
        <Text style={styles.label}>Correo electrónico</Text>
        <TextInput value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address"
          placeholder="tu@correo.com" style={styles.input} placeholderTextColor={COLORS.text3} />

        <Text style={styles.label}>Contraseña</Text>
        <TextInput value={password} onChangeText={setPassword} secureTextEntry
          placeholder="••••••••" style={styles.input} placeholderTextColor={COLORS.text3} />

        <TouchableOpacity onPress={submit} style={styles.btn}>
          <Text style={styles.btnText}>Iniciar sesión</Text>
        </TouchableOpacity>

        <Text style={styles.footer}>v6.0 · Funcionalidad offline disponible</Text>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 24, justifyContent: "center" },
  logo: { alignItems: "center", marginBottom: 40 },
  logoBox: { width: 60, height: 60, borderRadius: 14, backgroundColor: COLORS.accent, justifyContent: "center", alignItems: "center", marginBottom: 14 },
  logoText: { fontSize: 24, fontWeight: "900", color: "#fff" },
  title: { fontSize: 22, fontWeight: "700", color: COLORS.text1 },
  subtitle: { fontSize: 12, color: COLORS.text2, marginTop: 4 },
  form: { backgroundColor: COLORS.surface, padding: 24, borderRadius: 12, borderWidth: 1, borderColor: COLORS.border },
  label: { fontSize: 12, fontWeight: "600", color: COLORS.text1, marginBottom: 6, marginTop: 12 },
  input: { borderWidth: 1, borderColor: COLORS.border, borderRadius: 8, padding: 12, fontSize: 14, color: COLORS.text1 },
  btn: { backgroundColor: COLORS.accent, padding: 14, borderRadius: 8, alignItems: "center", marginTop: 20 },
  btnText: { color: "#fff", fontSize: 14, fontWeight: "700" },
  footer: { fontSize: 10, color: COLORS.text3, textAlign: "center", marginTop: 16 },
});

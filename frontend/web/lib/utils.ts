import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { format, formatDistanceToNow, isAfter, isBefore, addDays } from "date-fns";
import { es } from "date-fns/locale";

export function cn(...inputs: ClassValue[]) { return twMerge(clsx(inputs)); }
export function formatDate(date: string | Date) { return format(new Date(date), "dd/MM/yyyy", { locale: es }); }
export function formatDateTime(date: string | Date) { return format(new Date(date), "dd/MM/yyyy HH:mm", { locale: es }); }
export function timeAgo(date: string | Date) { return formatDistanceToNow(new Date(date), { addSuffix: true, locale: es }); }
export function isExpired(date: string | Date) { return isBefore(new Date(date), new Date()); }
export function isExpiringSoon(date: string | Date, days = 30) {
  const d = new Date(date);
  return isAfter(d, new Date()) && isBefore(d, addDays(new Date(), days));
}
export function formatCOP(amount: number) {
  return new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 }).format(amount);
}
export function scoreColor(score: number) {
  if (score >= 85) return "var(--green)";
  if (score >= 65) return "var(--amber)";
  return "var(--red)";
}
export function statusBadge(status: string): "green" | "amber" | "red" | "gray" {
  const map: Record<string, "green" | "amber" | "red" | "gray"> = {
    active: "green", ok: "green", completed: "green", paid: "green",
    warning: "amber", in_progress: "amber", pending: "amber", scheduled: "amber",
    critical: "red", open: "red", overdue: "red", suspended: "red",
    inactive: "gray", cancelled: "gray", archived: "gray",
  };
  return map[status] || "gray";
}

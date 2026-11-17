interface Dated {
  activity_date?: string | null;
}

const dayStart = (value: string | number | Date) => new Date(value).setHours(0, 0, 0, 0);

export function filterByPeriod<T extends Dated>(
  list: T[],
  filterType: string,
  customStartDate: string,
  customEndDate: string,
): T[] {
  const todayDate = new Date();
  todayDate.setHours(0, 0, 0, 0);
  return list.filter((item) => {
    if (!item.activity_date) return true;
    const itemDate = new Date(item.activity_date);
    itemDate.setHours(0, 0, 0, 0);
    if (filterType === 'this_week') {
      const startOfWeek = new Date(todayDate);
      startOfWeek.setDate(todayDate.getDate() - todayDate.getDay() + 1);
      const endOfWeek = new Date(startOfWeek);
      endOfWeek.setDate(startOfWeek.getDate() + 6);
      return itemDate >= startOfWeek && itemDate <= endOfWeek;
    }
    if (filterType === 'this_month') {
      return (
        itemDate.getMonth() === todayDate.getMonth() &&
        itemDate.getFullYear() === todayDate.getFullYear()
      );
    }
    if (filterType === 'custom') {
      const start = customStartDate ? new Date(customStartDate).setHours(0, 0, 0, 0) : null;
      const end = customEndDate ? new Date(customEndDate).setHours(23, 59, 59, 999) : null;
      if (start && itemDate.getTime() < start) return false;
      if (end && itemDate.getTime() > end) return false;
    }
    return true;
  });
}

// Tanpa tanggal dianggap aktif; tanggal sebelum hari ini masuk riwayat.
export function splitActiveHistory<T extends Dated>(list: T[]): { active: T[]; history: T[] } {
  const today = dayStart(new Date());
  return {
    active: list.filter((item) => !item.activity_date || dayStart(item.activity_date) >= today),
    history: list.filter((item) => !!item.activity_date && dayStart(item.activity_date) < today),
  };
}

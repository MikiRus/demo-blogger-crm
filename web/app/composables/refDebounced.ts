// Отложенная копия ref: поиск не дёргает API на каждую букву.
export function refDebounced<T>(source: Ref<T>, ms: number): Ref<T> {
  const out = ref(source.value) as Ref<T>
  let timer: ReturnType<typeof setTimeout> | undefined
  watch(source, (v) => {
    clearTimeout(timer)
    timer = setTimeout(() => (out.value = v), ms)
  })
  onScopeDispose(() => clearTimeout(timer))
  return out
}

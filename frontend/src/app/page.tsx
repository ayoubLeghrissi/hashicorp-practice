import { fetchProducts } from "./actions";
import PageClient from "./PageClient";

export const dynamic = 'force-dynamic';

export default async function Home() {
  let products = [];
  try {
    const res: any = await fetchProducts();
    products = res.products || [];
  } catch (error) {
    console.error("Failed to fetch products:", error);
  }

  return <PageClient initialProducts={products} userId="user-123" />;
}

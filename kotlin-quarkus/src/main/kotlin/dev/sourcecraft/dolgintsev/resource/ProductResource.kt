package dev.sourcecraft.dolgintsev.resource

import jakarta.transaction.Transactional
import jakarta.ws.rs.*
import jakarta.ws.rs.core.MediaType
import jakarta.ws.rs.core.Response
import dev.sourcecraft.dolgintsev.entity.Product

@Path("/api/products")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
class ProductResource {

    @GET
    fun getAllProducts(): List<Product> {
        return Product.listAll()
    }

    @GET
    @Path("/{id}")
    fun getProductById(@PathParam("id") id: Long): Response {
        val product = Product.findById(id)
        return if (product != null) {
            Response.ok(product).build()
        } else {
            Response.status(Response.Status.NOT_FOUND).build()
        }
    }

    @POST
    @Transactional
    fun createProduct(product: Product): Response {
        product.persist()
        return Response.status(Response.Status.CREATED).entity(product).build()
    }

    @PUT
    @Path("/{id}")
    @Transactional
    fun updateProduct(@PathParam("id") id: Long, updatedProduct: Product): Response {
        val product = Product.findById(id)
        return if (product != null) {
            product.name = updatedProduct.name
            product.description = updatedProduct.description
            product.price = updatedProduct.price
            product.quantity = updatedProduct.quantity
            Response.ok(product).build()
        } else {
            Response.status(Response.Status.NOT_FOUND).build()
        }
    }

    @DELETE
    @Path("/{id}")
    @Transactional
    fun deleteProduct(@PathParam("id") id: Long): Response {
        val deleted = Product.deleteById(id)
        return if (deleted) {
            Response.noContent().build()
        } else {
            Response.status(Response.Status.NOT_FOUND).build()
        }
    }
}

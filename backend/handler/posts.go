package handler

import (
	"context"

	"blog_backend/api"
	"blog_backend/model"

	"github.com/gosimple/slug"
)

func (h *Handler) GetPosts(ctx context.Context, request api.GetPostsRequestObject) (api.GetPostsResponseObject, error) {
	var posts []model.Post
	result := h.db.Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	response := make([]api.Post, 0, len(posts))
	for _, p := range posts {
		post := api.Post{
			Id:        &p.ID,
			Title:     &p.Title,
			Slug:      &p.Slug,
			Content:   &p.Content,
			CreatedAt: &p.CreatedAt,
		}
		response = append(response, post)
	}

	return api.GetPosts200JSONResponse(response), nil
}

func (h *Handler) GetPostsSlug(ctx context.Context, request api.GetPostsSlugRequestObject) (api.GetPostsSlugResponseObject, error) {
	var post model.Post
	result := h.db.Where("slug = ?", request.Slug).First(&post)
	if result.Error != nil {
		return api.GetPostsSlug404Response{}, nil
	}

	return api.GetPostsSlug200JSONResponse(api.Post{
		Id:        &post.ID,
		Title:     &post.Title,
		Slug:      &post.Slug,
		Content:   &post.Content,
		CreatedAt: &post.CreatedAt,
	}), nil
}

func (h *Handler) PostPosts(ctx context.Context, request api.PostPostsRequestObject) (api.PostPostsResponseObject, error) {
	slug := slug.Make(request.Body.Title)

	post := model.Post{
		Title:   request.Body.Title,
		Slug:    slug,
		Content: request.Body.Content,
	}

	result := h.db.Create(&post)
	if result.Error != nil {
		return nil, result.Error
	}

	id := int(post.ID)
	return api.PostPosts201JSONResponse(api.Post{
		Id:        &id,
		Title:     &post.Title,
		Slug:      &post.Slug,
		Content:   &post.Content,
		CreatedAt: &post.CreatedAt,
	}), nil
}

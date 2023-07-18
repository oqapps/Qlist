#include <iostream>
#include <wx/wxprec.h>
#include <wx/filedlg.h>
#include <wx/dataview.h>
#ifndef WX_PRECOMP
#include <wx/wx.h>
#endif

using namespace std;
class Node;
using NodePtr = unique_ptr<Node>;
using NodePtrArray = vector<NodePtr>;

class Node
{
public:
    Node(Node *parent,
         const wxString &key, const wxString &type,
         const wxString &value)
    {
        m_parent = parent;
        m_key = key;
        m_type = type;
        m_value = value;

        m_container = false;
    }

    Node(Node *parent,
         const wxString &branch)
    {
        m_parent = parent;
        m_key = "N/A";
        m_type = "N/A";
        m_value = "N/A";

        m_container = true;
    }

    ~Node() = default;

    bool IsContainer() const
    {
        return m_container;
    }

    Node *GetParent()
    {
        return m_parent;
    }
    NodePtrArray &GetChildren()
    {
        return m_children;
    }
    Node *GetNthChild(unsigned int n)
    {
        return m_children.at(n).get();
    }
    void Insert(Node *child, unsigned int n)
    {
        m_children.insert(m_children.begin() + n, NodePtr(child));
    }
    void Append(Node *child)
    {
        m_children.push_back(NodePtr(child));
    }
    unsigned int GetChildCount() const
    {
        return m_children.size();
    }

public: // public to avoid getters/setters
    wxString m_key;
    wxString m_type;
    wxString m_value;

    bool m_container;

private:
    Node *m_parent;
    NodePtrArray m_children;
};

class TreeModel : public wxDataViewModel
{
public:
    virtual void GetValue(wxVariant &variant,
                          const wxDataViewItem &item, unsigned int col) const override;
    virtual bool SetValue(const wxVariant &variant,
                          const wxDataViewItem &item, unsigned int col) override;

    virtual wxDataViewItem GetParent(const wxDataViewItem &item) const override;
    virtual bool IsContainer(const wxDataViewItem &item) const override;
    virtual unsigned int GetChildren(const wxDataViewItem &parent,
                                     wxDataViewItemArray &array) const override;
    virtual Node Add();

private:
    Node *root;
};

void TreeModel::GetValue(wxVariant &variant,
                         const wxDataViewItem &item, unsigned int col) const
{
    wxASSERT(item.IsOk());

    variant = "HI!";
}

Node TreeModel::Add()
{
    Node *n = new Node(root, "Classical music");
    return *n;
}

bool TreeModel::SetValue(const wxVariant &variant,
                         const wxDataViewItem &item, unsigned int col)
{
    // wxASSERT(item.IsOk());

    return true;
}

wxDataViewItem TreeModel::GetParent(const wxDataViewItem &item) const
{
    // the invisible root node has no parent
    if (!item.IsOk())
        return wxDataViewItem(0);

    /*Node *node = (Node*) item.GetID();

    // "MyMusic" also has no parent
    if (node == m_root)
        return wxDataViewItem(0);

    return wxDataViewItem( (void*) node->GetParent() );*/
    return wxDataViewItem(0);
}

bool TreeModel::IsContainer(const wxDataViewItem &item) const
{
    // the invisible root node can have children
    // (in our model always "MyMusic")
    if (!item.IsOk())
        return true;

    return false;
}

unsigned int TreeModel::GetChildren(const wxDataViewItem &parent,
                                    wxDataViewItemArray &array) const
{
    /*Node *node = (Node*) parent.GetID();
    if (!node)
    {
        array.Add( wxDataViewItem( (void*) m_root ) );
        return 1;
    }

    if (node->GetChildCount() == 0)
    {
        return 0;
    }

    for ( const auto& child : node->GetChildren() )
    {
        array.Add( wxDataViewItem( child.get() ) );
    }

    return array.size();*/
    return 0;
}

class MyApp : public wxApp
{
public:
    virtual bool OnInit();
};
class Frame : public wxFrame
{
public:
    Frame(const wxString &title, const wxPoint &pos, const wxSize &size);

private:
    void OnFileOpen(wxCommandEvent &event);
    void OnExit(wxCommandEvent &event);
    void OnAbout(wxCommandEvent &event);
    wxDataViewCtrl *tree;
    wxDECLARE_EVENT_TABLE();
};
enum
{
    ID_FILE = 1,
    ID_NEW = 2
};
wxBEGIN_EVENT_TABLE(Frame, wxFrame)
    EVT_MENU(ID_FILE, Frame::OnFileOpen)
        EVT_MENU(wxID_EXIT, Frame::OnExit)
            EVT_MENU(wxID_ABOUT, Frame::OnAbout)
                wxEND_EVENT_TABLE()
                    wxIMPLEMENT_APP(MyApp);
bool MyApp::OnInit()
{
    Frame *frame = new Frame("Qlist", wxPoint(50, 50), wxSize(550, 450));
    frame->Show(true);
    return true;
}
Frame::Frame(const wxString &title, const wxPoint &pos, const wxSize &size)
    : wxFrame(NULL, wxID_ANY, title, pos, size)
{
    wxMenu *menuFile = new wxMenu;
    menuFile->Append(ID_FILE, "&Open\tCtrl-O");
    menuFile->Append(ID_NEW, "&New\tCtrl-N");
    menuFile->Append(wxID_EXIT);
    menuFile->Append(wxID_ABOUT, "&About Qlist");
    menuFile->Append(wxID_PREFERENCES, "&Settings");
    wxMenuBar *menuBar = new wxMenuBar;
    menuBar->Append(menuFile, "&File");
    SetMenuBar(menuBar);
    wxDataViewCtrl *dataview = new wxDataViewCtrl(this, wxID_ANY, wxDefaultPosition, wxDefaultSize, wxDV_ROW_LINES);
    wxDataViewTextRenderer *tr =
        new wxDataViewTextRenderer("string", wxDATAVIEW_CELL_INERT);
    wxDataViewColumn *keycol =
        new wxDataViewColumn("Key", tr, 0, 180);
    wxDataViewColumn *typecol =
        new wxDataViewColumn("Type", tr, 0, 100);
    wxDataViewColumn *valuecol =
        new wxDataViewColumn("Value", tr, 0, 180);
    dataview->AppendColumn(keycol);
    dataview->AppendColumn(typecol);
    dataview->AppendColumn(valuecol);
    TreeModel *model = new TreeModel;
    model->Add();
    wxDataViewItem *i = new wxDataViewItem;
    dataview->AssociateModel(model);
}
void Frame::OnExit(wxCommandEvent &event)
{
    Close(true);
}
void Frame::OnAbout(wxCommandEvent &event)
{
    wxMessageBox("This is a wxWidgets' Hello world sample",
                 "About Hello World", wxOK | wxICON_INFORMATION);
}
void Frame::OnFileOpen(wxCommandEvent &event)
{
    wxFileDialog
        openFileDialog(this, "", "", "",
                       "Property-List files (*.plist)|*.plist", wxFD_OPEN | wxFD_FILE_MUST_EXIST);
    if (openFileDialog.ShowModal() == wxID_CANCEL)
        return;
}